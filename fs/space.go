package fs

import (
	"io/ioutil"
	"path/filepath"
	"syscall"

	"github.com/filecoin-project/specs-actors/actors/abi"
	"golang.org/x/xerrors"
)

// reserve reserves storage for the sector. `path` is the path of the directory containing sectors
func (f *FS) reserve(typ DataType, path StoragePath, qtyBytesNeeded uint64) error {
	f.lk.Lock()
	defer f.lk.Unlock()

	avail, fsavail, err := f.availableBytes(path)
	if err != nil {
		return err
	}

	if int64(qtyBytesNeeded) > avail {
		return xerrors.Errorf("not enough space in '%s', need %dB, available %dB (fs: %dB, reserved: %dB)",
			f.paths,
			qtyBytesNeeded,
			avail,
			fsavail,
			f.reservedBytes(path))
	}

	if _, ok := f.reserved[path]; !ok {
		f.reserved[path] = map[DataType]uint64{}
	}
	f.reserved[path][typ] += qtyBytesNeeded

	return nil
}

func (f *FS) Release(path SectorPath, sectorSize abi.SectorSize) {
	f.lk.Lock()
	defer f.lk.Unlock()

	f.reserved[path.storage()][path.typ()] -= overheadMul[path.typ()] * uint64(sectorSize)
}

func (f *FS) List(path StoragePath, typ DataType) ([]SectorPath, error) {
	tp := filepath.Join(string(path), string(typ))

	ents, err := ioutil.ReadDir(tp)
	if err != nil {
		return nil, err
	}

	out := make([]SectorPath, len(ents))
	for i, ent := range ents {
		out[i] = SectorPath(filepath.Join(tp, ent.Name()))
	}

	return out, nil
}

func (f *FS) reservedBytes(path StoragePath) int64 {
	var out int64
	rsvs, ok := f.reserved[path]
	if !ok {
		return 0
	}
	for _, r := range rsvs {
		out += int64(r)
	}
	return out
}

func (f *FS) availableBytes(path StoragePath) (int64, int64, error) {
	var fsstat syscall.Statfs_t

	if err := syscall.Statfs(string(path), &fsstat); err != nil {
		return 0, 0, err
	}

	fsavail := int64(fsstat.Bavail) * int64(fsstat.Bsize)

	avail := fsavail - f.reservedBytes(path)

	return avail, fsavail, nil
}
