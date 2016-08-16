/*
  This file is used to point the compiler to the actual opencl.h of the system.
  It is also used to check the version of opencl installed
*/

#ifdef __APPLE__
	#include <OpenCL/OpenCL.h>
#else

	#include "../../headers/ocl12/cl.h"
	#include "../../headers/ocl12/cl_gl.h"
	#include "../../headers/ocl12/cl_gl_ext.h"
	#include "../../headers/ocl12/cl_ext.h"
	#include "../../headers/ocl12/cl_egl.h"
/*
	#include <CL/opencl.h>
*/
#ifdef __WIN32
	#include "../../headers/ocl12/cl_d3d10.h"
	#include "../../headers/ocl12/cl_d3d11.h"
	#include "../../headers/ocl12/cl_dx9_media_sharing.h"
/*
	#include <CL/cl_dx9_media_sharing.h>
	#include <CL/cl_d3d10.h>
	#include <CL/cl_d3d11.h>
*/
#endif
#endif

#ifndef CL_VERSION_1_2
	#error "This package requires OpenCL 1.2"
#endif

#ifndef __gl_h_
#define GL_TEXTURE_1D                     0x0DE0
#define GL_TEXTURE_2D                     0x0DE1
#endif
#ifndef __glext_h_
#define GL_TEXTURE_3D                     0x806F
#define GL_TEXTURE_RECTANGLE              0x84F5
#define GL_TEXTURE_CUBE_MAP_POSITIVE_X    0x8515
#define GL_TEXTURE_CUBE_MAP_NEGATIVE_X    0x8516
#define GL_TEXTURE_CUBE_MAP_POSITIVE_Y    0x8517
#define GL_TEXTURE_CUBE_MAP_NEGATIVE_Y    0x8518
#define GL_TEXTURE_CUBE_MAP_POSITIVE_Z    0x8519
#define GL_TEXTURE_CUBE_MAP_NEGATIVE_Z    0x851A
#define GL_TEXTURE_1D_ARRAY               0x8C18
#define GL_TEXTURE_2D_ARRAY               0x8C1A
#define GL_TEXTURE_BUFFER                 0x8C2A
#endif
