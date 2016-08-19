package cl

/*
#include "./opencl.h"
*/
import "C"

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

//////////////// Basic Types ////////////////
type LocalMemType int

const (
	LocalMemTypeNone   LocalMemType = C.CL_NONE
	LocalMemTypeGlobal LocalMemType = C.CL_GLOBAL
	LocalMemTypeLocal  LocalMemType = C.CL_LOCAL
)

var localMemTypeMap = map[LocalMemType]string{
	LocalMemTypeNone:   "None",
	LocalMemTypeGlobal: "Global",
	LocalMemTypeLocal:  "Local",
}

func (t LocalMemType) String() string {
	name := localMemTypeMap[t]
	if name == "" {
		name = "Unknown"
	}
	return name
}

type MemCacheType int

const (
	MemCacheTypeNone           MemCacheType = C.CL_NONE
	MemCacheTypeReadOnlyCache  MemCacheType = C.CL_READ_ONLY_CACHE
	MemCacheTypeReadWriteCache MemCacheType = C.CL_READ_WRITE_CACHE
)

func (ct MemCacheType) String() string {
	switch ct {
	case MemCacheTypeNone:
		return "None"
	case MemCacheTypeReadOnlyCache:
		return "ReadOnly"
	case MemCacheTypeReadWriteCache:
		return "ReadWrite"
	}
	return fmt.Sprintf("Unknown(%x)", int(ct))
}

type MemFlag int

const (
	MemReadWrite    MemFlag = C.CL_MEM_READ_WRITE
	MemWriteOnly    MemFlag = C.CL_MEM_WRITE_ONLY
	MemReadOnly     MemFlag = C.CL_MEM_READ_ONLY
	MemUseHostPtr   MemFlag = C.CL_MEM_USE_HOST_PTR
	MemAllocHostPtr MemFlag = C.CL_MEM_ALLOC_HOST_PTR
	MemCopyHostPtr  MemFlag = C.CL_MEM_COPY_HOST_PTR
)

type MemObjectType int

const (
	MemObjectTypeBuffer  MemObjectType = C.CL_MEM_OBJECT_BUFFER
)

type MapFlag int

const (
	// This flag specifies that the region being mapped in the memory object is being mapped for reading.
	MapFlagRead                  MapFlag = C.CL_MAP_READ
	MapFlagWrite                 MapFlag = C.CL_MAP_WRITE
)

func (mf MapFlag) toCl() C.cl_map_flags {
	return C.cl_map_flags(mf)
}

type MappedMemObject struct {
	ptr        unsafe.Pointer
	size       int
	rowPitch   int
	slicePitch int
}

//////////////// Abstract Types ////////////////
type MemObject struct {
        clMem C.cl_mem
        size  int
}

//////////////// Basic Functions ////////////////
func retainMemObject(b *MemObject) {
        if b.clMem != nil {
                C.clRetainMemObject(b.clMem)
        }
}

func releaseMemObject(b *MemObject) {
        if b.clMem != nil {
                C.clReleaseMemObject(b.clMem)
                b.clMem = nil
        }
}

func newMemObject(mo C.cl_mem, size int) *MemObject {
        memObject := &MemObject{clMem: mo, size: size}
        runtime.SetFinalizer(memObject, releaseMemObject)
        return memObject
}


//////////////// Abstract Functions ////////////////
func (mb *MappedMemObject) ByteSlice() []byte {
	var byteSlice []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&byteSlice))
	sliceHeader.Cap = mb.size
	sliceHeader.Len = mb.size
	sliceHeader.Data = uintptr(mb.ptr)
	return byteSlice
}

func (mb *MappedMemObject) Ptr() unsafe.Pointer {
	return mb.ptr
}

func (mb *MappedMemObject) Size() int {
	return mb.size
}

func (mb *MappedMemObject) RowPitch() int {
	return mb.rowPitch
}

func (mb *MappedMemObject) SlicePitch() int {
	return mb.slicePitch
}

func (b *MemObject) Retain() {
        retainMemObject(b)
}

func (b *MemObject) Release() {
        releaseMemObject(b)
}

func (b *MemObject) GetType() (string, error) {
	if b.clMem != nil {
		var tmp C.cl_mem_object_type
		err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_TYPE, C.size_t(unsafe.Sizeof(tmp)), unsafe.Pointer(&tmp), nil)
		if toError(err) != nil {
			return "Unknown", toError(err)
		}
		switch {
		case MemObjectType(tmp) == MemObjectTypeBuffer:
			return "Buffer", nil
		case MemObjectType(tmp) == MemObjectTypeImage2D:
			return "Image2D", nil
		case MemObjectType(tmp) == MemObjectTypeImage3D:
			return "Image3D", nil
		default:
			return "Unknown", nil
		}
	}
	return "Unknown", toError(C.CL_INVALID_MEM_OBJECT)
}

func (b *MemObject) GetContext() (*Context, error) {
        if b.clMem != nil {
		var tmp C.cl_context
		err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_CONTEXT, C.size_t(unsafe.Sizeof(tmp)), unsafe.Pointer(&tmp), nil)
		if toError(err) != nil {
			return nil, nil
		}
		return &Context{clContext: tmp, devices: nil}, toError(err)
        }
        return nil, toError(C.CL_INVALID_MEM_OBJECT)
}

func (b *MemObject) GetSize() (int, error) {
        if b.clMem != nil {
		var tmp C.size_t
		err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_SIZE, C.size_t(unsafe.Sizeof(tmp)), unsafe.Pointer(&tmp), nil)
		if toError(err) != nil {
			return int(-1), toError(err)
		}
		return int(tmp), nil
        }
        return 0, toError(C.CL_INVALID_MEM_OBJECT)
}

func (b *MemObject) GetRefenceCount() (int, error) {
        if b.clMem != nil {
		var tmp C.cl_uint
		err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_REFERENCE_COUNT, C.size_t(unsafe.Sizeof(tmp)), unsafe.Pointer(&tmp), nil)
		if toError(err) != nil {
			return 0, toError(err)
		}
		return int(tmp), nil
        }
        return 0, toError(C.CL_INVALID_MEM_OBJECT)
}

func (b *MemObject) GetMapCount() (int, error) {
        if b.clMem != nil {
	        var tmp C.cl_uint
        	err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_MAP_COUNT, C.size_t(unsafe.Sizeof(tmp)), unsafe.Pointer(&tmp), nil)
	        if toError(err) != nil {
        	        return 0, toError(err)
	        }
        	return int(tmp), nil
        }
        return 0, toError(C.CL_INVALID_MEM_OBJECT)
}

func (b *MemObject) GetHostPtr() (unsafe.Pointer, error) {
	if b.clMem != nil {
		var tmp unsafe.Pointer
		err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_HOST_PTR, C.size_t(unsafe.Sizeof(tmp)), unsafe.Pointer(&tmp), nil)
	        if toError(err) != nil {
        	        return nil, toError(err)
	        }
        	return tmp, nil
        }
        return nil, toError(C.CL_INVALID_MEM_OBJECT)
}

func (b *MemObject) GetFlags() (MemFlag, error) {
        if b.clMem != nil {
		var tmp C.cl_mem_flags
		err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_FLAGS, C.size_t(unsafe.Sizeof(tmp)), unsafe.Pointer(&tmp), nil)
	        if toError(err) != nil {
        	        return -1, toError(err)
	        }
		switch {
		case tmp == C.CL_MEM_READ_WRITE:
			return MemReadWrite, nil
	        case tmp == C.CL_MEM_WRITE_ONLY:
        	        return MemWriteOnly, nil
	        case tmp == C.CL_MEM_READ_ONLY:
        	        return MemReadOnly, nil
	        case tmp == C.CL_MEM_USE_HOST_PTR:
        	        return MemUseHostPtr, nil
	        case tmp == C.CL_MEM_ALLOC_HOST_PTR:
        	        return MemAllocHostPtr, nil
	        case tmp == C.CL_MEM_COPY_HOST_PTR:
        	        return MemCopyHostPtr, nil
		default:
	        	return -1, nil
	 	}
        }
        return -1, toError(C.CL_INVALID_MEM_OBJECT)
}

// Enqueues a command to map a region of the buffer object given by buffer into the host address space and returns a pointer to this mapped region.
func (q *CommandQueue) EnqueueMapBuffer(buffer *MemObject, blocking bool, flags MapFlag, offset, size int, eventWaitList []*Event) (*MappedMemObject, *Event, error) {
	var event C.cl_event
	var err C.cl_int
	ptr := C.clEnqueueMapBuffer(q.clQueue, buffer.clMem, clBool(blocking), flags.toCl(), C.size_t(offset), C.size_t(size), C.cl_uint(len(eventWaitList)), eventListPtr(eventWaitList), &event, &err)
	if err != C.CL_SUCCESS {
		return nil, nil, toError(err)
	}
	ev := newEvent(event)
	if ptr == nil {
		return nil, ev, ErrUnknown
	}
	return &MappedMemObject{ptr: ptr, size: size}, ev, nil
}

// Enqueues a command to unmap a previously mapped region of a memory object.
func (q *CommandQueue) EnqueueUnmapMemObject(buffer *MemObject, mappedObj *MappedMemObject, eventWaitList []*Event) (*Event, error) {
	var event C.cl_event
	if err := C.clEnqueueUnmapMemObject(q.clQueue, buffer.clMem, mappedObj.ptr, C.cl_uint(len(eventWaitList)), eventListPtr(eventWaitList), &event); err != C.CL_SUCCESS {
		return nil, toError(err)
	}
	return newEvent(event), nil
}

// Enqueues a command to copy a buffer object to another buffer object.
func (q *CommandQueue) EnqueueCopyBuffer(srcBuffer, dstBuffer *MemObject, srcOffset, dstOffset, byteCount int, eventWaitList []*Event) (*Event, error) {
	var event C.cl_event
	err := toError(C.clEnqueueCopyBuffer(q.clQueue, srcBuffer.clMem, dstBuffer.clMem, C.size_t(srcOffset), C.size_t(dstOffset), C.size_t(byteCount), C.cl_uint(len(eventWaitList)), eventListPtr(eventWaitList), &event))
	return newEvent(event), err
}

// Enqueue commands to write to a buffer object from host memory.
func (q *CommandQueue) EnqueueWriteBuffer(buffer *MemObject, blocking bool, offset, dataSize int, dataPtr unsafe.Pointer, eventWaitList []*Event) (*Event, error) {
	var event C.cl_event
	err := toError(C.clEnqueueWriteBuffer(q.clQueue, buffer.clMem, clBool(blocking), C.size_t(offset), C.size_t(dataSize), dataPtr, C.cl_uint(len(eventWaitList)), eventListPtr(eventWaitList), &event))
	return newEvent(event), err
}

func (q *CommandQueue) EnqueueWriteBufferByte(buffer *MemObject, blocking bool, offset int, data []byte, eventWaitList []*Event) (*Event, error) {
	dataPtr := unsafe.Pointer(&data[0])
	dataSize := int(unsafe.Sizeof(data[0])) * len(data)
	return q.EnqueueWriteBuffer(buffer, blocking, offset, dataSize, dataPtr, eventWaitList)
}

func (q *CommandQueue) EnqueueWriteBufferFloat32(buffer *MemObject, blocking bool, offset int, data []float32, eventWaitList []*Event) (*Event, error) {
	dataPtr := unsafe.Pointer(&data[0])
	dataSize := int(unsafe.Sizeof(data[0])) * len(data)
	return q.EnqueueWriteBuffer(buffer, blocking, offset, dataSize, dataPtr, eventWaitList)
}

// Enqueue commands to read from a buffer object to host memory.
func (q *CommandQueue) EnqueueReadBuffer(buffer *MemObject, blocking bool, offset, dataSize int, dataPtr unsafe.Pointer, eventWaitList []*Event) (*Event, error) {
	var event C.cl_event
	err := toError(C.clEnqueueReadBuffer(q.clQueue, buffer.clMem, clBool(blocking), C.size_t(offset), C.size_t(dataSize), dataPtr, C.cl_uint(len(eventWaitList)), eventListPtr(eventWaitList), &event))
	return newEvent(event), err
}

func (q *CommandQueue) EnqueueReadBufferByte(buffer *MemObject, blocking bool, offset int, data []byte, eventWaitList []*Event) (*Event, error) {
	dataPtr := unsafe.Pointer(&data[0])
	dataSize := int(unsafe.Sizeof(data[0])) * len(data)
	return q.EnqueueReadBuffer(buffer, blocking, offset, dataSize, dataPtr, eventWaitList)
}

func (q *CommandQueue) EnqueueReadBufferFloat32(buffer *MemObject, blocking bool, offset int, data []float32, eventWaitList []*Event) (*Event, error) {
	dataPtr := unsafe.Pointer(&data[0])
	dataSize := int(unsafe.Sizeof(data[0])) * len(data)
	return q.EnqueueReadBuffer(buffer, blocking, offset, dataSize, dataPtr, eventWaitList)
}

func (ctx *Context) CreateBufferUnsafe(flags MemFlag, size int, dataPtr unsafe.Pointer) (*MemObject, error) {
        var err C.cl_int
        clBuffer := C.clCreateBuffer(ctx.clContext, C.cl_mem_flags(flags), C.size_t(size), dataPtr, &err)
        if err != C.CL_SUCCESS {
                return nil, toError(err)
        }
        if clBuffer == nil {
                return nil, ErrUnknown
        }
        return newMemObject(clBuffer, size), nil
}

func (ctx *Context) CreateEmptyBuffer(flags MemFlag, size int) (*MemObject, error) {
        return ctx.CreateBufferUnsafe(flags, size, nil)
}

func (ctx *Context) CreateEmptyBufferFloat32(flags MemFlag, size int) (*MemObject, error) {
        return ctx.CreateBufferUnsafe(flags, 4*size, nil)
}

func (ctx *Context) CreateBuffer(flags MemFlag, data []byte) (*MemObject, error) {
        return ctx.CreateBufferUnsafe(flags, len(data), unsafe.Pointer(&data[0]))
}

//float64
func (ctx *Context) CreateBufferFloat32(flags MemFlag, data []float32) (*MemObject, error) {
        return ctx.CreateBufferUnsafe(flags, 4*len(data), unsafe.Pointer(&data[0]))
}
