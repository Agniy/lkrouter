package controler

import (
	"fmt"
	"github.com/gin-gonic/gin"
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

func IfRoomRecordStatusIsStopped(room string) bool {
	recordStatus, err := redisdb.GetRoomRecordStatus(room)
	if err != nil {
		fmt.Println("Error getting record status from redis", err)
		return true
	}
	return recordStatus == "stopped"
}

func StartEgressController(c *gin.Context) {
	data := EgressStartData{}
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	//check if room record status is stopped
	if !IfRoomRecordStatusIsStopped(data.Room) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Room record status is not stopped"})
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
		"rec": true,
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

	response := EgressStartResponse{
		Room:     data.Room,
		Status:   "ok",
		EgressID: egressId,
	}
	c.JSON(http.StatusAccepted, &response)
}
