// +build halv1

package wrapper

// #cgo CFLAGS: -I${SRCDIR}/../lora_gateway/libloragw/inc
// #cgo LDFLAGS: -lm ${SRCDIR}/../lora_gateway/libloragw/libloragw.a
// #include "config.h"
// #include "loragw_hal.h"
// #include "loragw_gps.h"
import "C"
import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/pkg/errors"
)

var gps *os.File

var gpsTimeReference = C.struct_tref{}
var gpsTimeReferenceMutex = &sync.Mutex{}

var validCoordinates bool
var coordinates GPSCoordinates
var coordinatesMutex = &sync.Mutex{}

var gpsMaxAge = time.Second * 30

const bufferSize = 128

func GetGPSCoordinates() (GPSCoordinates, error) {
	coordinatesMutex.Lock()
	defer coordinatesMutex.Unlock()
	tmpCoordinates := coordinates

	if !validCoordinates {
		return tmpCoordinates, errors.New("No valid coordinates obtained from GPS yet")
	}

	if !gpsActive() {
		return tmpCoordinates, errors.New("GPS not active")
	}

	return tmpCoordinates, nil
}

// timeReference returns the GPS time reference in a time.Time format
func timeReference() time.Time {
	gpsTimeReferenceMutex.Lock()
	currentTimeReference := gpsTimeReference
	gpsTimeReferenceMutex.Unlock()
	return time.Unix(int64(currentTimeReference.systime), 0)
}

// LoRaGPSEnable acts as a wrapper for lgw_gps_enable
func LoRaGPSEnable(TTYPath string) error {
	fd := C.int(0)

	// HAL only supports u-blox7 for now, so gps_family must be "ubx7"
	ok := (C.lgw_gps_enable(C.CString(TTYPath), C.CString("ubx7"), C.speed_t(0), &fd) == C.LGW_GPS_SUCCESS)
	if !ok {
		return errors.New("Failed GPS configuration - impossible to open port for GPS sync (check permissions?)")
	}

	gps = os.NewFile(uintptr(fd), "GPS")

	return nil
}

func gpsActive() bool {
	return gps != nil
}

func checkGPSTimeReference() bool {
	if !gpsActive() {
		return false
	}

	if timeReference().Add(gpsMaxAge).Before(time.Now()) {
		// GPS Time Reference considered obsolete
		return false
	}

	return true
}

func UpdateGPSData(ctx log.Interface) error {
	var (
		coord    C.struct_coord_s
		coordErr C.struct_coord_s
		ts       C.uint32_t
		utcTime  C.struct_timespec
	)
	reader := bufio.NewReader(gps)
	buffer, err := reader.ReadBytes('\x0a')
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "GPS interface read error")
	}

	gpsRawData := string(buffer[:])

	nmea := C.lgw_parse_nmea(C.CString(gpsRawData), C.int(cap(buffer)))
	if nmea != C.NMEA_RMC {
		// No sync to do
		ctx.Debug("Unknown GPS status")
		return nil
	}

	ctx.Debug("Recommended Minimum sentence C received, triggering GPS sync")
	if C.lgw_gps_get(&utcTime, nil, nil) != C.LGW_GPS_SUCCESS {
		ctx.Debug("Couldn't get UTC time from GPS")
		return nil
	}

	ctx.Debug("Fetching GPS timestamp")
	concentratorMutex.Lock()
	ok := C.lgw_get_trigcnt(&ts) == C.LGW_GPS_SUCCESS
	concentratorMutex.Unlock()

	if !ok {
		ctx.Warn("Failed to read concentrator timestamp")
		return nil
	}

	ctx.Debug("Fetching GPS time reference")
	gpsTimeReferenceMutex.Lock()
	ok = C.lgw_gps_sync(&gpsTimeReference, ts, utcTime) == C.LGW_GPS_SUCCESS
	gpsTimeReferenceMutex.Unlock()

	if !ok {
		ctx.Warn("GPS out of sync, keeping previous time reference")
		return nil
	}
	ctx.WithField("GPSDateComputation", timeReference()).Debug("Date sync with GPS complete")

	ctx.Debug("Fetching GPS coordinates")
	coordinatesMutex.Lock()
	ok = C.lgw_gps_get(nil, &coord, &coordErr) == C.LGW_GPS_SUCCESS
	// For the moment, coordErr is unused, because the back-end doesn't handle the GPS's margin of error.
	// One possible improvement, if it is handled upstream, would be handling this.
	if !ok {
		ctx.Warn("Couldn't retrieve GPS coordinates")
		return nil
	}

	coordinates = GPSCoordinates{
		Altitude:  float64(coord.alt),
		Latitude:  float64(coord.lat),
		Longitude: float64(coord.lon),
	}
	validCoordinates = true
	ctx.WithFields(log.Fields{"Altitude": coordinates.Altitude, "Latitude": coordinates.Latitude, "Longitude": coordinates.Longitude}).Info("GPS coordinates updated")
	coordinatesMutex.Unlock()
	return nil
}
