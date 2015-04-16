package rexray

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emccode/rexray/storagedriver"
)

var (
	debug          string
	storageDrivers string
	drivers        map[string]storagedriver.Driver
)
var (
	ErrDriverBlockDeviceDiscovery = errors.New("Driver Block Device discovery failed")
	ErrDriverInstanceDiscovery    = errors.New("Driver Instance discovery failed")
	ErrDriverVolumeDiscovery      = errors.New("Driver Volume discovery failed")
	ErrDriverSnapshotDiscovery    = errors.New("Driver Snapshot discovery failed")
	ErrMultipleDriversDetected    = errors.New("Multiple drivers detected, must declare with driver with env of REXRAY_STORAGEDRIVER=")
)

func init() {
	debug = strings.ToUpper(os.Getenv("REXRAY_DEBUG"))
	storageDrivers = strings.ToLower(os.Getenv("REXRAY_STORAGEDRIVERS"))
	var err error
	drivers, err = storagedriver.GetDrivers(storageDrivers)
	if err != nil && debug == "TRUE" {
		fmt.Println(err)
	}
	if len(drivers) == 0 {
		log.Fatalf("No drivers initialized")
	}
}

// GetBlockDeviceMapping performs storage introspection and
// returns a listing of block devices from the guest
func GetBlockDeviceMapping() ([]*storagedriver.BlockDevice, error) {
	var allBlockDevices []*storagedriver.BlockDevice
	for _, driver := range drivers {
		blockDevices, err := driver.GetBlockDeviceMapping()
		if err != nil {
			return []*storagedriver.BlockDevice{}, fmt.Errorf("Error: %s: %s", ErrDriverBlockDeviceDiscovery, err)
		}

		if len(blockDevices.([]*storagedriver.BlockDevice)) > 0 {
			for _, blockDevice := range blockDevices.([]*storagedriver.BlockDevice) {
				allBlockDevices = append(allBlockDevices, blockDevice)
			}
		}
	}

	return allBlockDevices, nil

}

func GetInstance() ([]*storagedriver.Instance, error) {
	var allInstances []*storagedriver.Instance
	for _, driver := range drivers {
		instance, err := driver.GetInstance()
		if err != nil {
			return []*storagedriver.Instance{}, fmt.Errorf("Error: %s: %s", ErrDriverInstanceDiscovery, err)
		}

		allInstances = append(allInstances, instance.(*storagedriver.Instance))

	}

	return allInstances, nil
}

func GetVolume(volumeID string) ([]*storagedriver.Volume, error) {
	var allVolumes []*storagedriver.Volume

	for _, driver := range drivers {
		volumes, err := driver.GetVolume(volumeID)
		if err != nil {
			return []*storagedriver.Volume{}, fmt.Errorf("Error: %s: %s", ErrDriverVolumeDiscovery, err)
		}

		if len(volumes.([]*storagedriver.Volume)) > 0 {
			for _, volume := range volumes.([]*storagedriver.Volume) {
				allVolumes = append(allVolumes, volume)
			}
		}
	}
	return allVolumes, nil
}

func GetSnapshot(volumeID, snapshotID string) ([]*storagedriver.Snapshot, error) {
	var allSnapshots []*storagedriver.Snapshot

	for _, driver := range drivers {
		snapshots, err := driver.GetSnapshot(volumeID, snapshotID)
		if err != nil {
			return []*storagedriver.Snapshot{}, fmt.Errorf("Error: %s: %s", ErrDriverSnapshotDiscovery, err)
		}

		if len(snapshots.([]*storagedriver.Snapshot)) > 0 {
			for _, snapshot := range snapshots.([]*storagedriver.Snapshot) {
				allSnapshots = append(allSnapshots, snapshot)
			}
		}
	}
	return allSnapshots, nil
}

func CreateSnapshot(runAsync bool, volumeID, description string) ([]*storagedriver.Snapshot, error) {
	if len(drivers) > 1 {
		return []*storagedriver.Snapshot{}, ErrMultipleDriversDetected
	}
	for _, driver := range drivers {
		snapshot, err := driver.CreateSnapshot(runAsync, volumeID, description)
		if err != nil {
			return []*storagedriver.Snapshot{}, err
		}
		return snapshot.([]*storagedriver.Snapshot), nil
	}
	return []*storagedriver.Snapshot{}, nil
}

func RemoveSnapshot(snapshotID string) error {
	if len(drivers) > 1 {
		return ErrMultipleDriversDetected
	}
	for _, driver := range drivers {
		err := driver.RemoveSnapshot(snapshotID)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateVolume(runAsync bool, snapshotID string, volumeType string, IOPS int64, size int64) (*storagedriver.Volume, error) {
	if len(drivers) > 1 {
		return &storagedriver.Volume{}, ErrMultipleDriversDetected
	}
	for _, driver := range drivers {
		volume, err := driver.CreateVolume(runAsync, snapshotID, volumeType, IOPS, size)
		if err != nil {
			return &storagedriver.Volume{}, err
		}
		return volume.(*storagedriver.Volume), nil
	}
	return &storagedriver.Volume{}, nil
}

func RemoveVolume(volumeID string) error {
	if len(drivers) > 1 {
		return ErrMultipleDriversDetected
	}
	for _, driver := range drivers {
		err := driver.RemoveVolume(volumeID)
		if err != nil {
			return err
		}
	}
	return nil
}

func AttachVolume(runAsync bool, volumeID string, instanceID string) ([]*storagedriver.VolumeAttachment, error) {
	if len(drivers) > 1 {
		return []*storagedriver.VolumeAttachment{}, ErrMultipleDriversDetected
	}
	for _, driver := range drivers {
		if instanceID == "" {
			instance, err := driver.GetInstance()
			if err != nil {
				return []*storagedriver.VolumeAttachment{}, fmt.Errorf("Error: %s: %s", ErrDriverInstanceDiscovery, err)
			}
			instanceID = instance.(*storagedriver.Instance).InstanceID
		}

		volumeAttachment, err := driver.AttachVolume(runAsync, volumeID, instanceID)
		if err != nil {
			return []*storagedriver.VolumeAttachment{}, err
		}
		return volumeAttachment.([]*storagedriver.VolumeAttachment), nil
	}
	return []*storagedriver.VolumeAttachment{}, nil
}

func DetachVolume(runAsync bool, volumeID string, instanceID string) error {
	if len(drivers) > 1 {
		return ErrMultipleDriversDetected
	}
	for _, driver := range drivers {
		if instanceID == "" {
			instance, err := driver.GetInstance()
			if err != nil {
				fmt.Errorf("Error: %s: %s", ErrDriverInstanceDiscovery, err)
			}
			instanceID = instance.(*storagedriver.Instance).InstanceID
		}

		err := driver.DetachVolume(runAsync, volumeID, instanceID)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
