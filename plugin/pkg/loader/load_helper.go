//     Copyright (C) 2020-2021, IrineSistiana
//
//     This file is part of mosdns.
//
//     mosdns is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     mosdns is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with this program.  If not, see <https://www.gnu.org/licenses/>.

package loader

import (
	"fmt"
	"strings"
	"time"

	"github.com/ca17/datahub/plugin/pkg/v2data"
	"github.com/golang/protobuf/proto"
)

var matcherCache = NewCache()

const (
	cacheTTL = time.Second * 30
)

// mustHaveAttr checks if attr has all wanted attrs.
func mustHaveAttr(attr, wanted []string) bool {
	if len(wanted) == 0 {
		return true
	}
	if len(attr) == 0 {
		return false
	}

	for _, w := range wanted {
		ok := false
		for _, got := range attr {
			if got == w {
				ok = true
				break
			}
		}
		if !ok { // this attr is not in d.
			return false
		}
	}
	return true
}

func LoadGeoSiteFromDAT(file, countryCode string) (*v2data.GeoSite, error) {
	geoSiteList, err := LoadGeoSiteList(file)
	if err != nil {
		return nil, err
	}

	countryCode = strings.ToUpper(countryCode)
	entry := geoSiteList.GetEntry()
	for i := range entry {
		if strings.ToUpper(entry[i].CountryCode) == countryCode {
			return entry[i], nil
		}
	}

	return nil, fmt.Errorf("can not find category %s in %s", countryCode, file)
}

func LoadGeoSiteList(file string) (*v2data.GeoSiteList, error) {
	data, raw, err := matcherCache.LoadFromCacheOrRawDisk(file)
	if err != nil {
		return nil, err
	}
	// load from cache
	if geoSiteList, ok := data.(*v2data.GeoSiteList); ok {
		return geoSiteList, nil
	}

	// load from disk
	geoSiteList := new(v2data.GeoSiteList)
	if err := proto.Unmarshal(raw, geoSiteList); err != nil {
		return nil, err
	}

	return geoSiteList, nil
}


func LoadGeoIPListFromDAT(file string) (*v2data.GeoIPList, error) {
	data, raw, err := matcherCache.LoadFromCacheOrRawDisk(file)
	if err != nil {
		return nil, err
	}
	// load from cache
	if geoIPList, ok := data.(*v2data.GeoIPList); ok {
		return geoIPList, nil
	}

	// load from disk
	geoIPList := new(v2data.GeoIPList)
	if err := proto.Unmarshal(raw, geoIPList); err != nil {
		return nil, err
	}

	// cache the file
	matcherCache.Put(file, geoIPList, cacheTTL)
	return geoIPList, nil
}
