package egresserv

import (
	"context"
	"fmt"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"lkrouter/config"
)

func StartTrackEgress(roomName string) string {
	config := config.GetConfig()
	ctx := context.Background()
	egressClient := lksdk.NewEgressClient(
		config.LVHost,
		config.LVApiKey,
		config.LVApiSecret,
	)

	fileRequest := &livekit.RoomCompositeEgressRequest{
		RoomName:  roomName,
		Layout:    "1x1",
		AudioOnly: true,
		Output: &livekit.RoomCompositeEgressRequest_File{
			File: &livekit.EncodedFileOutput{
				FileType: livekit.EncodedFileType_OGG,
				Filepath: "livekit-demo/track-test.ogg",
				Output: &livekit.EncodedFileOutput_S3{
					S3: &livekit.S3Upload{
						AccessKey: config.AWSAccessKey,
						Secret:    config.AWSSecret,
						Region:    config.AWSRegion,
						Bucket:    config.AWSBucket,
					},
				},
			},
		},
	}

	info, err := egressClient.StartRoomCompositeEgress(ctx, fileRequest)
	if err != nil {
		fmt.Println("Error in StartTrackEgress:", err)
	}
	fmt.Println("StartTrackEgress:", info.EgressId)

	return info.EgressId
}
