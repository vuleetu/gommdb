package gommdb

/*
#cgo LDFLAGS: -lmaxminddb
#include "maxminddb.h"
#include <stdio.h>
#include <stdlib.h>

double entry_data_double_value(MMDB_entry_data_s* data) {
    return data->double_value;
}
*/
import "C"

import (
    "errors"
    "unsafe"
)

var (
    LONGITUDE_NOT_FOUND = errors.New("Longitude not found")
    LATITUDE_NOT_FOUND = errors.New("Latitude not found")
    LOCATION_NOT_FOUND = errors.New("Location not found")
)

type MMDB struct {
    mmdb *C.MMDB_s
}

type Location struct {
    Lat, Lng float64
}

func New(db string) (*MMDB, error) {
    var mmdb C.MMDB_s
    cdb := C.CString(db)
    defer func() {
        C.free(unsafe.Pointer(cdb))
    }()

    status := C.MMDB_open(cdb, C.MMDB_MODE_MMAP, &mmdb)
    if C.MMDB_SUCCESS != status {
        return nil, errors.New(C.GoString(C.MMDB_strerror(status)))
    }

    return &MMDB{&mmdb}, nil
}

func (db *MMDB) Close() {
    C.MMDB_close(db.mmdb)
}

func (db *MMDB) Location(ip string) (*Location, error) {
    cip := C.CString(ip)
    defer func() {
        C.free(unsafe.Pointer(cip))
    }()

    var gai_error, mmdb_error C.int
    result := C.MMDB_lookup_string(db.mmdb, cip, &gai_error, &mmdb_error)
    if result.found_entry {
        var latitude, longitude C.double
        var entry_data C.MMDB_entry_data_s

        var slocation = C.CString("location")
        var slatitude = C.CString("latitude")
        var slongitude = C.CString("longitude")

        defer func() {
            C.free(unsafe.Pointer(slocation))
            C.free(unsafe.Pointer(slatitude))
            C.free(unsafe.Pointer(slongitude))
        }()

        var path [3]*C.char
        path[0] = slocation
        path[1] = slatitude
        path[2] = nil

        status := C.MMDB_aget_value(&result.entry, &entry_data, &path[0])
        if C.MMDB_SUCCESS != status {
            return nil, errors.New(C.GoString(C.MMDB_strerror(status)))
        }

        if entry_data._type != C.MMDB_DATA_TYPE_DOUBLE {
            return nil, errors.New("Invalid location type")
        }

        if entry_data.has_data {
            latitude = C.entry_data_double_value(&entry_data)
        } else {
            return nil, LATITUDE_NOT_FOUND
        }

        path[1] = slongitude
        status = C.MMDB_aget_value(&result.entry, &entry_data, &path[0])
        if C.MMDB_SUCCESS != status {
            return nil, errors.New(C.GoString(C.MMDB_strerror(status)))
        }

        if entry_data._type != C.MMDB_DATA_TYPE_DOUBLE {
            return nil, errors.New("Invalid location type")
        }

        if entry_data.has_data {
            longitude = C.entry_data_double_value(&entry_data)
        } else {
            return nil, LONGITUDE_NOT_FOUND
        }

        return &Location{float64(latitude), float64(longitude)}, nil
    }

    return nil, LONGITUDE_NOT_FOUND
}
