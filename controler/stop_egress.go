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
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// get egress ID from redis
	egressId, err := redisdb.Get("room_egress_" + data.Room)
	if err != nil {
		fmt.Println("Error getting egress ID from redis", err)
		c.AbortWithError(http.StatusBadRequest, err)
	}

	eggressInfo, err := egresserv.StopTrackEgress(egressId.(string))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	// save to redis db only if get status is ending
	if eggressInfo.Status == livekit.EgressStatus_EGRESS_ENDING {
		err = redisdb.SetRoomRecordStatus(data.Room, "stopping", 1*time.Minute)
		if err != nil {
			fmt.Println("Error saving egress ID to redis", err)
		}
	}

	// remove from redis db
	err = redisdb.Del("room_egress_" + data.Room)
	if err != nil {
		fmt.Println("Error removing egress ID from redis", err)
	}

	fmt.Println("Try to send to room: ", data.Room, " new metadata: \"rec\": \"false\"")

	// update room metadata
	room, err := livekitserv.NewLiveKitService().UpdateRoomMData(data.Room, map[string]interface{}{
		"rec":        false,
		"rec-status": "stopping",
	})
	if err != nil {
		fmt.Println("Error updating room metadata", err)
	}

	fmt.Println("Room metadata updated", room.Metadata)

	// Update room metadata in MongoDB
	err = mrequests.SetRecordStatus(data.Room, false)
	if err != nil {
		fmt.Println("Error updating room metadata in MongoDB", err)
	}

	awslogs.AddSLog(map[string]string{
		"func":    "StopEgressController",
		"message": "Record successful stopped, egressId: " + egressId.(string),
		"room":    data.Room,
	})

	response := EgressStopResponse{
		Room:     data.Room,
		Status:   "ok",
		EgressID: egressId.(string),
	}
	c.JSON(http.StatusAccepted, &response)
}
