// Copyright (C) 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vulkan

import (
	"context"

	"github.com/google/gapid/core/log"
	"github.com/google/gapid/gapis/api"
	"github.com/google/gapid/gapis/api/transform"
)

// wireframe returns a transform that set all the graphics pipeline to be
// created with rasterization polygon mode == VK_POLYGON_MODE_LINE
func wireframe(ctx context.Context) transform.Transformer {
	ctx = log.Enter(ctx, "Wirefrmae")
	return transform.Transform("Wireframe", func(ctx context.Context,
		id api.CmdID, cmd api.Cmd, out transform.Writer) {
		s := out.State()
		l := s.MemoryLayout
		cb := CommandBuilder{Thread: cmd.Thread()}
		cmd.Extras().Observations().ApplyReads(s.Memory.ApplicationPool())
		switch cmd := cmd.(type) {
		case *VkCreateGraphicsPipelines:
			count := uint64(cmd.CreateInfoCount)
			infos := cmd.PCreateInfos.Slice(0, count, l)
			newInfos := make([]VkGraphicsPipelineCreateInfo, count)
			newRasterStateDatas := make([]api.AllocResult, count)
			for i := uint64(0); i < count; i++ {
				info := infos.Index(i, l).Read(ctx, cmd, s, nil)
				rasterState := info.PRasterizationState.Read(ctx, cmd, s, nil)
				rasterState.PolygonMode = VkPolygonMode_VK_POLYGON_MODE_LINE
				newRasterStateDatas[i] = s.AllocDataOrPanic(ctx, rasterState)
				info.PRasterizationState = NewVkPipelineRasterizationStateCreateInfoᶜᵖ(newRasterStateDatas[i].Ptr())
				newInfos[i] = info
			}
			newInfosData := s.AllocDataOrPanic(ctx, newInfos)
			newCmd := cb.VkCreateGraphicsPipelines(cmd.Device,
				cmd.PipelineCache, cmd.CreateInfoCount, newInfosData.Ptr(),
				cmd.PAllocator, cmd.PPipelines, cmd.Result).AddRead(newInfosData.Data())
			for _, r := range newRasterStateDatas {
				newCmd.AddRead(r.Data())
			}
			for _, w := range cmd.Extras().Observations().Writes {
				newCmd.AddWrite(w.Range, w.ID)
			}
			out.MutateAndWrite(ctx, id, newCmd)
		case *RecreateGraphicsPipeline:
			info := cmd.PCreateInfo.Read(ctx, cmd, s, nil)
			rasterState := info.PRasterizationState.Read(ctx, cmd, s, nil)
			rasterState.PolygonMode = VkPolygonMode_VK_POLYGON_MODE_LINE
			newRasterStateData := s.AllocDataOrPanic(ctx, rasterState)
			info.PRasterizationState = NewVkPipelineRasterizationStateCreateInfoᶜᵖ(newRasterStateData.Ptr())
			newInfoData := s.AllocDataOrPanic(ctx, info)
			newCmd := cb.RecreateGraphicsPipeline(cmd.Device, cmd.PipelineCache,
				newInfoData.Ptr(), cmd.PPipeline).AddRead(newInfoData.Data()).AddRead(
				newRasterStateData.Data())
			for _, w := range cmd.Extras().Observations().Writes {
				newCmd.AddWrite(w.Range, w.ID)
			}
			out.MutateAndWrite(ctx, id, newCmd)
		default:
			out.MutateAndWrite(ctx, id, cmd)
		}
	})
}
