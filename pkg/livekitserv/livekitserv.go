package livekitserv

import (
	"context"
	"encoding/json"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/sirupsen/logrus"
	"lkrouter/config"
	"log"
	"os"
	"strings"
)

func NewLiveKitService() *LiveKitService {
	cfg := config.GetConfig()
	return &LiveKitService{
		client: lksdk.NewRoomServiceClient(cfg.LVHost, cfg.LVApiKey, cfg.LVApiSecret),
		logger: logrus.New(),
	}
}

type LiveKitService struct {
	client *lksdk.RoomServiceClient
	logger *logrus.Logger
}

func (l *LiveKitService) UpdateUserMData(roomID string, identity string, metadata map[string]interface{}) (*livekit.ParticipantInfo, error) {
	ctx := context.Background()
	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	req := &livekit.UpdateParticipantRequest{
		Room:     roomID,
		Identity: identity,
		Metadata: string(jsonBytes),
	}
	participant, err := l.client.UpdateParticipant(ctx, req)
	if err != nil {
		return nil, err
	}
	return participant, nil
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

func (l *LiveKitService) GetAllActiveCalls() (*livekit.ListRoomsResponse, error) {
	ctx := context.Background()
	req := &livekit.ListRoomsRequest{}
	rooms, err := l.client.ListRooms(ctx, req)
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

func (l *LiveKitService) GetAudioTrackID(roomID string, participantID string) (string, error) {
	res, err := l.client.GetParticipant(context.Background(), &livekit.RoomParticipantIdentity{
		Room:     roomID,
		Identity: participantID,
	})

	if err != nil {
		return "", err
	}

	for _, track := range res.Tracks {
		if track.Type == livekit.TrackType_AUDIO {
			return track.Sid, nil
		}
	}

	return "", nil
}

func (l *LiveKitService) DeleteRoom(roomID string) error {
	_, err := l.client.DeleteRoom(context.Background(), &livekit.DeleteRoomRequest{
		Room: roomID,
	})
	return err
}

func (l *LiveKitService) SendMessageToParticipants(roomID string, message []byte, topicName string) error {
	sendResponse, err := l.client.SendData(context.Background(), &livekit.SendDataRequest{
		Room:  roomID,
		Data:  message,
		Topic: &topicName,
	})
	if err != nil {
		return err
	}

	l.logger.Infof("Send message to participant %s in room %s, response: %v", roomID, sendResponse)

	return nil
}

func (l *LiveKitService) GetRoomParticipants(roomID string) (*livekit.ListParticipantsResponse, error) {
	ctx := context.Background()
	req := &livekit.ListParticipantsRequest{
		Room: roomID,
	}
	roomParticipants, err := l.client.ListParticipants(ctx, req)
	if err != nil {
		return nil, err
	}

	return roomParticipants, nil
}

func (l *LiveKitService) RealParticipantsByRoom(roomID string) ([]*livekit.ParticipantInfo, error) {
	logger := log.New(os.Stdout, "ListParticipantsByRoom: ", 0)

	roomResponse, err := l.GetRoomParticipants(roomID)
	if err != nil {
		logger.Println("Error when try ListParticipantsByRoom: ", err)
		return nil, err
	}

	// If there are no real participants in the room, close it.
	activeParticipants := []*livekit.ParticipantInfo{}
	for _, participant := range roomResponse.Participants {
		if participant.Identity != "chat-agent" && !strings.HasPrefix(participant.Identity, "EG_") {
			activeParticipants = append(activeParticipants, participant)
		}
	}

	return activeParticipants, nil
}
