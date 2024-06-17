package controler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"lkrouter/domain"
	"lkrouter/pkg/awslogs"
	"lkrouter/pkg/egresserv"
	"lkrouter/pkg/livekitserv"
	"lkrouter/pkg/mongodb/mrequests"
	"lkrouter/pkg/redisdb"
	"net/http"
	"time"
)

type EgressStartData struct {
	Room    string `json:"room"`
	Company string `json:"company"`
}

type EgressStartResponse struct {
	Room     string `json:"room"`
	Status   string `json:"status"`
	EgressID string `json:"egressID"`
}

func IfRoomRecordStatusIsStopping(room string) bool {
	recordStatus, err := redisdb.GetRoomRecordStatus(room)
	if err != nil {
		fmt.Println("Error getting record status from redis", err)
		return false
	}
	return recordStatus == "stopping"
}

func StartEgressController(c *gin.Context) {
	data := EgressStartData{}
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	//check if room record status is stopping
	if IfRoomRecordStatusIsStopping(data.Room) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Room record status is not stopped",
			"notifications": []domain.RoomHttpNotification{
				{
					MsgCode:  "RECORD_NOT_READY_ERROR",
					Type:     "error",
					Head:     "Record is processing, please wait",
					Msg:      "Processing previous rec part, please try again later(if record is big more time needed)",
					Infinite: true,
				},
			},
		})
		return
	}

	egressId, err := egresserv.StartTrackEgress(data.Room, data.Company)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	// save to redis db
	err = redisdb.Set("room_egress_"+data.Room, egressId, 24*time.Hour)
	if err != nil {
		fmt.Println("Error saving egress ID to redis", err)
	}

	fmt.Println("Try to send to room: ", data.Room, " new metadata: \"rec\": \"true\"")

	room, err := livekitserv.NewLiveKitService().UpdateRoomMData(data.Room, map[string]interface{}{
		"rec":        true,
		"rec-status": "started",
	})

	if err != nil {
		fmt.Println("Error updating room metadata", err)
	}

	// Update room metadata in MongoDB
	err = mrequests.SetRecordStatus(data.Room, true)
	if err != nil {
		fmt.Println("Error updating room metadata in MongoDB", err)
	}

	fmt.Println("Room metadata updated", room.Metadata)

	awslogs.AddSLog(map[string]string{
		"func":    "StartEgressController",
		"message": "Record successful started, egressId: " + egressId,
		"room":    data.Room,
	})

	response := EgressStartResponse{
		Room:     data.Room,
		Status:   "ok",
		EgressID: egressId,
	}
	c.JSON(http.StatusAccepted, &response)
}
