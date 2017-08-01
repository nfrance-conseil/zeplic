package test

import (
	"fmt"
	"time"

	"github.com/nfrance-conseil/zeplic/lib"
	"testing"
)
/*
func TestDelete(t *testing.T) {
	// Snapshots list
	dataset := "tank/test"
	SnapshotsList := []string{"abcd-1234:tank/test@SNAP_2016-August-24_16:00:00", "abcd-1234:tank/test@SNAP_2016-August-25_16:00:00", "abcd-1234:tank/test@SNAP_2016-August-26_16:00:00", "abcd-1234:tank/test@SNAP_2016-September-24_16:00:00", "abcd-1234:tank/test@SNAP_2016-September-25_16:00:00", "abcd-1234:tank/test@SNAP_2016-September-26_16:00:00", "abcd-1234:tank/test@SNAP_2016-October-24_16:00:00", "abcd-1234:tank/test@SNAP_2016-October-25_16:00:00", "abcd-1234:tank/test@SNAP_2016-October-26_16:00:00", "abcd-1234:tank/test@SNAP_2016-November-24_16:00:00", "abcd-1234:tank/test@SNAP_2016-November-25_16:00:00", "abcd-1234:tank/test@SNAP_2016-November-26_16:00:00", "abcd-1234:tank/test@SNAP_2016-December-24_16:00:00", "abcd-1234:tank/test@SNAP_2016-December-25_16:00:00", "abcd-1234:tank/test@SNAP_2016-December-26_16:00:00", "abcd-1234:tank/test@SNAP_2017-January-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-January-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-January-26_16:00:00", "abcd-1234:tank/test@SNAP_2017-February-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-February-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-February-26_16:00:00", "abcd-1234:tank/test@SNAP_2017-March-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-March-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-March-26_16:00:00", "abcd-1234:tank/test@SNAP_2017-April-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-April-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-April-26_16:00:00", "abcd-1234:tank/test@SNAP_2017-May-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-May-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-May-26_16:00:00", "abcd-1234:tank/test@SNAP_2017-June-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-June-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-June-26_16:00:00", "abcd-1234:tank/test@SNAP_2017-June-28_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-05_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-08_11:00:00", "abcd-1234:tank/test@SNAP_2017-July-09_12:00:00", "abcd-1234:tank/test@SNAP_2017-July-12_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-17_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-18_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-19_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-19_18:00:00", "abcd-1234:tank/test@SNAP_2017-July-20_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-21_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-22_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_11:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_12:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_13:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_14:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_15:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_11:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_12:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_13:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_14:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_15:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_11:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_12:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_13:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_14:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_15:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_17:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_18:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_19:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_20:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_21:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_22:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_23:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_00:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_01:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_02:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_03:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_04:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_05:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_06:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_07:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_08:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_09:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_10:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_11:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_12:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_13:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_14:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_14:30:00", "abcd-1234:tank/test@SNAP_2017-July-26_15:00:00", "abcd-1234:tank/test@SNAP_2017-July-26_15:30:00", "abcd-1234:tank/test@SNAP_2017-July-26_16:00:00"}
	prefix := "SNAP"
	retention := "24d1w1m1y"
	destroy, toDestroy := lib.Delete(dataset, SnapshotsList, prefix, retention)
	if destroy == false {
		t.Errorf("Delete() test failed!")
	} else {
		toDestroyList := []string{"abcd-1234:tank/test@SNAP_2017-July-25_17:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_18:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_15:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_14:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_13:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_12:00:00", "abcd-1234:tank/test@SNAP_2017-July-25_11:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_15:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_14:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_13:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_12:00:00", "abcd-1234:tank/test@SNAP_2017-July-24_11:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_15:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_14:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_13:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_12:00:00", "abcd-1234:tank/test@SNAP_2017-July-23_11:00:00", "abcd-1234:tank/test@SNAP_2017-July-18_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-17_16:00:00", "abcd-1234:tank/test@SNAP_2017-July-09_12:00:00", "abcd-1234:tank/test@SNAP_2017-July-08_11:00:00", "abcd-1234:tank/test@SNAP_2017-June-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-June-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-May-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-May-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-April-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-April-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-March-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-March-24_16:00:00", "abcd-1234:tank/test@SNAP_2017-February-25_16:00:00", "abcd-1234:tank/test@SNAP_2017-February-24_16:00:00", "abcd-1234:tank/test@SNAP_2016-December-25_16:00:00", "abcd-1234:tank/test@SNAP_2016-December-24_16:00:00", "abcd-1234:tank/test@SNAP_2016-November-25_16:00:00", "abcd-1234:tank/test@SNAP_2016-November-24_16:00:00", "abcd-1234:tank/test@SNAP_2016-October-25_16:00:00", "abcd-1234:tank/test@SNAP_2016-October-24_16:00:00", "abcd-1234:tank/test@SNAP_2016-September-25_16:00:00", "abcd-1234:tank/test@SNAP_2016-September-24_16:00:00", "abcd-1234:tank/test@SNAP_2016-August-25_16:00:00"}
		for i := 0; i < len(toDestroy); i++ {
			if toDestroy[i] != toDestroyList[i] {
				t.Errorf("Delete() test failed!")
			}
		}
	}
}
*/
func TestNewSnapshot(t *testing.T) {
	// Snapshots list, cron format and prefix
	SnapshotsList := []string{"1a2b3c4d-5678-efgh-6789-0z1a2b3c4d5e:tank/test@SNAP_2017-July-01_12:00:00", "1a2b3c4d-5678-efgh-6789-0z1a2b3c4d5e:tank/test@SNAP_2017-July-01_11:00:00#sent"}
	prefix := "SNAP"
	creation := "0 * * * *"
	take, SnapshotName := lib.NewSnapshot(SnapshotsList, creation, prefix)

	// Actual time
	year, month, day := time.Now().Date()
	hour, _, _ := time.Now().Clock()

	// Name of snapshot
	name := fmt.Sprintf("%s_%d-%s-%02d_%02d:00:00", prefix, year, month, day, hour)

	// Check
	if take == false || SnapshotName != name {
		t.Errorf("NewSnapshot() test failed!")
	}
}

func TestSend(t *testing.T) {
	// Snapshots list and policy to sync
	dataset := "tank/test"
	SnapshotsList := []string{"1a2b3c4d-5678-efgh-6789-0z1a2b3c4d5e:tank/test@SNAP_2017-July-01_22:00:00", "1a2b3c4d-5678-efgh-6789-0z1a2b3c4d5e:tank/test@SNAP_2017-July-01_11:00:00#sent"}
	prefix := "SNAP"
	SyncPolicy := "asap"
	send, SnapshotUUID := lib.Send(dataset, SnapshotsList, SyncPolicy, prefix)
	if send == false || SnapshotUUID != "1a2b3c4d-5678-efgh-6789-0z1a2b3c4d5e" {
		t.Errorf("Send() test failed!")
		t.Errorf("%s", SnapshotUUID)
	}

	// Check with cron format
	SyncPolicy = "0 23 * * 2"
	send, SnapshotUUID = lib.Send(dataset, SnapshotsList, SyncPolicy, prefix)
	if send == false || SnapshotUUID != "1a2b3c4d-5678-efgh-6789-0z1a2b3c4d5e" {
		t.Errorf("Send() test failed!")
		t.Errorf("%s", SnapshotUUID)
	}
}
