package controler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/livekit/protocol/livekit"
	"lkrouter/pkg/awslogs"
	"lkrouter/pkg/egresserv"
	"lkrouter/pkg/livekitserv"
	"lkrouter/pkg/mongodb/mrequests"
	"lkrouter/pkg/redisdb"
	"net/http"
	"time"
)

type EgressStopData struct {
	Room string `json:"room"`
}

type EgressStopResponse struct {
	Room     string `json:"room"`
	Status   string `json:"status"`
	EgressID string `json:"egressID"`
}

func StopEgressController(c *gin.Context) {
	data := EgressStopData{}
	if err := c.BindJSON(&data); err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "StopEgressController",
			"message": fmt.Sprintf("Error binding json to EgressStopData: %v", err),
			"type":    awslogs.MsgTypeError,
		})

		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// get egress ID from redis
	egressId, err := redisdb.Get("room_egress_" + data.Room)
	if err != nil {

		msg := fmt.Sprintf("Redis error getting egress ID, room: %v, error: %v", data.Room, err)

		awslogs.AddSLog(map[string]string{
			"func":    "StopEgressController",
			"message": msg,
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})

		c.AbortWithError(http.StatusBadRequest, err)
	}

	eggressInfo, err := egresserv.StopTrackEgress(egressId.(string))
	if err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "StopEgressController",
			"message": fmt.Sprintf("Error stopping egress, room: %v, error: %v", data.Room, err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})

		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	// save to redis db only if get status is ending
	if eggressInfo.Status == livekit.EgressStatus_EGRESS_ENDING {
		err = redisdb.SetRoomRecordStatus(data.Room, "stopping", 1*time.Minute)
		if err != nil {

			msg := fmt.Sprintf("Redis error saving egress ID, room: %v, error: %v", data.Room, err)

			awslogs.AddSLog(map[string]string{
				"func":    "StopEgressController",
				"message": msg,
				"type":    awslogs.MsgTypeError,
				"room":    data.Room,
			})
		}
	}

	// remove from redis db
	err = redisdb.Del("room_egress_" + data.Room)
	if err != nil {

		msg := fmt.Sprintf("Redis error removing egress ID from redis, room: %v, error: %v", data.Room, err)

		awslogs.AddSLog(map[string]string{
			"func":    "StopEgressController",
			"message": msg,
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	// update room metadata
	_, err = livekitserv.NewLiveKitService().UpdateRoomMData(data.Room, map[string]interface{}{
		"rec":        false,
		"rec-status": "stopping",
	})
	if err != nil {

		msg := fmt.Sprintf("Livekit error updating room metadata, room: %v, error: %v", data.Room, err)

		awslogs.AddSLog(map[string]string{
			"func":    "StopEgressController",
			"message": msg,
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	// Update room metadata in MongoDB
	err = mrequests.SetRecordStatus(data.Room, false)
	if err != nil {
		msg := fmt.Sprintf("MongoDB error updating room metadata, room: %v, error: %v", data.Room, err)
		awslogs.AddSLog(map[string]string{
			"func":    "StopEgressController",
			"message": msg,
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	awslogs.AddSLog(map[string]string{
		"func":    "StopEgressController",
		"message": "Record successful stopped, egressId: " + egressId.(string),
		"room":    data.Room,
		"type":    awslogs.MsgTypeInfo,
	})

	response := EgressStopResponse{
		Room:     data.Room,
		Status:   "ok",
		EgressID: egressId.(string),
	}
	c.JSON(http.StatusAccepted, &response)
}
