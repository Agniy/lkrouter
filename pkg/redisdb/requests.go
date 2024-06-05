package redisdb

import "time"

const (
	ROOM_RECORD_STATUS = "room_record_status_"
)

func GetRoomRecordStatus(room string) (string, error) {
	recordStatus, err := Get(ROOM_RECORD_STATUS + room)
	if err != nil {
		return "", err
	}
	return recordStatus.(string), nil
}

func SetRoomRecordStatus(room string, status string, timeout time.Duration) error {
	return Set(ROOM_RECORD_STATUS+room, status, timeout)
}

func DelRoomRecordStatus(room string) error {
	return Del(ROOM_RECORD_STATUS + room)
}
