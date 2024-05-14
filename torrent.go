// torrent.go
package main

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
)

func extractNameFromMetaInfo(metaInfo *metainfo.MetaInfo) (string, error) {
	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return "", err
	}
	return info.Name, nil
}

func getInfoHash(metaInfo *metainfo.MetaInfo) string {
	infoHash := metaInfo.HashInfoBytes()
	return hex.EncodeToString(infoHash[:])
}

func getTotalSize(metaInfo *metainfo.MetaInfo) (int64, error) {
	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return 0, err
	}

	var totalSize int64
	for _, file := range info.Files {
		totalSize += file.Length
	}
	return totalSize, nil
}

func generateMagnetLink(metaInfo *metainfo.MetaInfo, fileName string, totalSize int64) string {
	var builder strings.Builder
	infoHash := metaInfo.HashInfoBytes()

	builder.WriteString(fmt.Sprintf("magnet:?xt=urn:btih:%s", hex.EncodeToString(infoHash.Bytes())))

	if fileName != "" {
		builder.WriteString("&dn=")
		builder.WriteString(url.QueryEscape(fileName))
	}

	builder.WriteString(fmt.Sprintf("&xl=%d", totalSize))

	return builder.String()
}

func modifyMetadata(metaInfo *metainfo.MetaInfo, createdBy, comment string) {
	if *verboseFlag {
		if len(metaInfo.AnnounceList) > 0 {
			fmt.Println("Modifying torrent metadata...")
			fmt.Println("Removing the following trackers:")
			for _, trackers := range metaInfo.AnnounceList {
				for _, tracker := range trackers {
					fmt.Println(" -", tracker)
				}
			}
		} else {
			fmt.Println("Modifying torrent metadata... (no trackers found)")
		}
	}

	metaInfo.Announce = ""
	metaInfo.AnnounceList = nil
	metaInfo.CreatedBy = createdBy
	metaInfo.Comment = comment
}
