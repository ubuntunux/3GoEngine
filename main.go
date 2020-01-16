package main

import (
	"log"
	"runtime"
	"time"

	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
	"github.com/xlab/closer"

	"github.com/ubuntunux/GoEngine3D/VulkanContext"
)

var appInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 0, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "VulkanDraw\x00",
	PEngineName:        "vulkango.com\x00",
}

func init() {
	runtime.LockOSThread()
}

func main() {
	procAddr := glfw.GetVulkanGetInstanceProcAddress()
	if procAddr == nil {
		panic("GetInstanceProcAddress is nil")
	}
	vk.SetGetInstanceProcAddr(procAddr)

	orPanic(glfw.Init())
	orPanic(vk.Init())
	defer closer.Close()

	var (
		v   VulkanContext.VulkanDeviceInfo
		s   VulkanContext.VulkanSwapchainInfo
		r   VulkanContext.VulkanRenderInfo
		b   VulkanContext.VulkanBufferInfo
		gfx VulkanContext.VulkanGfxPipelineInfo
	)

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err := glfw.CreateWindow(1024, 768, "Vulkan Info", nil, nil)
	orPanic(err)

	createSurface := func(instance interface{}) uintptr {
		surface, err := window.CreateWindowSurface(instance, nil)
		orPanic(err)
		return surface
	}

	v, err = VulkanContext.NewVulkanDevice(appInfo,
		window.GLFWWindow(),
		window.GetRequiredInstanceExtensions(),
		createSurface)
	orPanic(err)
	s, err = v.CreateSwapchain()
	orPanic(err)
	r, err = VulkanContext.CreateRenderer(v.Device, s.DisplayFormat)
	orPanic(err)
	err = s.CreateFramebuffers(r.RenderPass, nil)
	orPanic(err)
	b, err = v.CreateBuffers()
	orPanic(err)
	gfx, err = VulkanContext.CreateGraphicsPipeline(v.Device, s.DisplaySize, r.RenderPass)
	orPanic(err)
	log.Println("[INFO] swapchain lengths:", s.SwapchainLen)
	err = r.CreateCommandBuffers(s.DefaultSwapchainLen())
	orPanic(err)

	doneC := make(chan struct{}, 2)
	exitC := make(chan struct{}, 2)
	defer closer.Bind(func() {
		exitC <- struct{}{}
		<-doneC
		log.Println("Bye!")
	})
	VulkanContext.VulkanInit(&v, &s, &r, &b, &gfx)

	fpsTicker := time.NewTicker(time.Second / 60000)

	current_time := time.Now()
	for {
		now := time.Now()
		delta := now.Sub(current_time).Seconds()
		current_time = now
		if 0.0 < delta {
			log.Println(1.0 / delta)
		}

		select {
		case <-exitC:
			VulkanContext.DestroyInOrder(&v, &s, &r, &b, &gfx)
			window.Destroy()
			glfw.Terminate()
			fpsTicker.Stop()
			doneC <- struct{}{}
			return
		case <-fpsTicker.C:
			if window.ShouldClose() {
				exitC <- struct{}{}
				continue
			}
			glfw.PollEvents()
			VulkanContext.VulkanDrawFrame(v, s, r)
		}
	}
}

func orPanic(err interface{}) {
	switch v := err.(type) {
	case error:
		if v != nil {
			panic(err)
		}
	case vk.Result:
		if err := vk.Error(v); err != nil {
			panic(err)
		}
	case bool:
		if !v {
			panic("condition failed: != true")
		}
	}
}
