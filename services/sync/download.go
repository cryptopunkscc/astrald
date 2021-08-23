package sync

import (
	"errors"
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/repo"
	"io"
	"log"
)

func download(
	repoClient repo.LocalRepository,
	filesClient repo.RemoteRepository,
	id fid.ID,
) (err error) {
	var reader repo.Reader
	var writer repo.Writer

	// Check if file exist
	log.Println(Port, "getting files reader for id", id.String())
	if reader, err = repoClient.Reader(id); err != nil {
		log.Println(Port, "cannot obtain remote reader", err)
		return
	} else {
		_ = reader.Close()
	}

	// Obtain remote reader
	log.Println(Port, "getting files reader for id", id.String())
	if reader, err = filesClient.Reader(id); err != nil {
		log.Println(Port, "cannot obtain remote reader", err)
		return
	}

	// Obtain local writer
	log.Println(Port, "getting repo writer", id.String())
	if writer, err = repoClient.Writer(); err != nil {
		log.Println(Port, "cannot obtain local writer", err)
		return
	}

	// Write file size
	log.Println(Port, "writing file size", id.String())
	_, err = writer.WriteUInt32(uint32(id.Size))
	if err != nil {
		log.Println(Port, "cannot write file size", id.Size, id.String())
		return err
	}

	// Copy file into local file system
	defer func() { _ = reader.Close() }()
	defer func() { _, _ = writer.Finalize() }()
	log.Println(Port, "coping file to local repo file with size", id.Size, id.String())
	if l, err := io.CopyN(writer, reader, int64(id.Size)); err != nil {
		log.Println(Port, "cannot copy file", l, err)
		return err
	}

	// Obtain calculated id for a saved file
	log.Println(Port, "getting file id")
	id2, err := writer.Finalize()
	if err != nil {
		log.Println(Port, "cannot finalize file copy", err)
		return err
	}

	// Verify calculated id against the received
	log.Println(Port, "verifying ids")
	rid := id.String()
	cid := id2.String()
	if rid != cid {
		return errors.New("received id " + rid + " is different than calculated " + cid )
	}

	// Finish
	log.Println(Port, "finish coping file to local repo", id.String())
	return
}