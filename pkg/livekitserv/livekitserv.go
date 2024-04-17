package livekitserv

import (
	"context"
	"encoding/json"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"lkrouter/config"
)

func NewLiveKitService() *LiveKitService {
	cfg := config.GetConfig()
	return &LiveKitService{
		client: lksdk.NewRoomServiceClient(cfg.LVHost, cfg.LVApiKey, cfg.LVApiSecret),
	}
}

type LiveKitService struct {
	client *lksdk.RoomServiceClient
}

func (l *LiveKitService) UpdateRoomMData(roomID string, metadata map[string]interface{}) (*livekit.Room, error) {
	ctx := context.Background()
	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	req := &livekit.UpdateRoomMetadataRequest{
		Room:     roomID,
		Metadata: string(jsonBytes),
	}
	room, err := l.client.UpdateRoomMetadata(ctx, req)
	if err != nil {
		return nil, err
	}
	return room, nil
}
