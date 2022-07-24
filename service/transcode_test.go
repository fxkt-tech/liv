package echosrv

import (
	"context"
	"testing"
)

func TestSimpleTranscodeMP4(t *testing.T) {
	var (
		ctx    = context.Background()
		params = &TranscodeParams{
			Infile: "in1.mp4",
			Subs: []*SubTranscodeParams{
				{
					Outfile: "out1.mp4",
					Filters: &TranscodeFilters{
						Video: &TranscodeVideo{
							Height: 540,
						},
						Logo: []*TranscodeLogo{
							{
								File: "logo1.png",
								Dx:   10,
								Dy:   8,
							},
						},
					},
				},
				// {
				// 	Outfile: "out2.mp4",
				// 	Filters: &TranscodeFilters{
				// 		Video: &TranscodeVideo{
				// 			Height: 720,
				// 		},
				// 	},
				// },
			},
		}
	)
	SimpleTranscodeMP4(ctx, params)
}
