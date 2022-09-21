package differ

import (
	"errors"
	"folder-scan/internal"
	"folder-scan/util"
)

const (
	ChangeModeSame    = "same"
	ChangeModeNew     = "new"
	ChangeModeDeleted = "deleted"
	ChangeModeChanged = "changed"
)

type DataChangeInfo struct {
	Old *internal.FSData `json:"old,omitempty"`
	Cur *internal.FSData `json:"cur,omitempty"`

	ChangeMode string `json:"changeMode"`
}

type FileChangeInfo DataChangeInfo

func NewFileChangeInfo(old, cur *internal.FileInfo) *FileChangeInfo {
	changeMode := ChangeModeSame
	if old == nil && cur == nil {
		panic("Tried to diff between two nil objects")
	}
	if old != nil && cur != nil {
		if old.Size != cur.Size {
			changeMode = ChangeModeChanged
		}
	}
	if old == nil {
		changeMode = ChangeModeNew
		old = &internal.FileInfo{
			Name: cur.Name,
			Path: cur.Path,
		}
	}
	if cur == nil {
		changeMode = ChangeModeDeleted
		cur = &internal.FileInfo{
			Name: old.Name,
			Path: old.Path,
		}
	}
	return &FileChangeInfo{
		Old:        (*internal.FSData)(old),
		Cur:        (*internal.FSData)(cur),
		ChangeMode: changeMode,
	}
}

type FolderChangeInfo struct {
	Data *DataChangeInfo `json:"data,omitempty"`

	Subs  []*FolderChangeInfo `json:"subs,omitempty"`
	Files []*FileChangeInfo   `json:"files,omitempty"`
}

func (f *FolderChangeInfo) searchChanges() {
	if f.Data.ChangeMode == ChangeModeNew || f.Data.ChangeMode == ChangeModeDeleted {
		return
	}

	for _, sub := range f.Subs {
		sub.searchChanges()
		if sub.Data.ChangeMode != ChangeModeSame {
			f.Data.ChangeMode = ChangeModeChanged
		}
	}

	for _, file := range f.Files {
		if file.ChangeMode != ChangeModeSame {
			f.Data.ChangeMode = ChangeModeChanged
		}
	}
}

func NewFolderChangeInfo(old, cur *internal.FolderInfo) *FolderChangeInfo {
	if old == nil && cur == nil {
		panic("Tried to diff between two nil objects")
	}
	changeMode := ChangeModeSame
	if old == nil {
		changeMode = ChangeModeNew
		old = &internal.FolderInfo{
			Data: &internal.FSData{
				Name:   cur.Data.Name,
				Path:   cur.Data.Path,
				Parent: cur.Data.Parent,
			},
			FilesSize:      0,
			SubfoldersSize: 0,
			Subs:           make([]*internal.FolderInfo, 0),
			Files:          make([]*internal.FileInfo, 0),
		}
	}
	if cur == nil {
		changeMode = ChangeModeDeleted
		cur = &internal.FolderInfo{
			Data: &internal.FSData{
				Name:   old.Data.Name,
				Path:   old.Data.Path,
				Parent: old.Data.Parent,
			},
			FilesSize:      0,
			SubfoldersSize: 0,
			Subs:           make([]*internal.FolderInfo, 0),
			Files:          make([]*internal.FileInfo, 0),
		}
	}

	info := FolderChangeInfo{
		Data: &DataChangeInfo{
			Old:        old.Data,
			Cur:        cur.Data,
			ChangeMode: changeMode,
		},
		Subs:  make([]*FolderChangeInfo, 0, util.MaxInt(len(old.Subs), len(cur.Subs))),
		Files: make([]*FileChangeInfo, 0, util.MaxInt(len(old.Files), len(cur.Files))),
	}

	// to quickly get old subfolder with the same name as new one
	oldSubsMap := make(map[string]*internal.FolderInfo)
	// to find deleted subfolders (exist in old.subs but not in cur.subs)
	oldSubsUnusedSet := make(map[*internal.FolderInfo]bool, len(old.Subs))
	for _, sub := range old.Subs {
		oldSubsUnusedSet[sub] = true
		oldSubsMap[sub.Data.Name] = sub
	}

	// all changed/identical/new are covered
	for _, curSub := range cur.Subs {
		oldSub := oldSubsMap[curSub.Data.Name]
		info.Subs = append(info.Subs, NewFolderChangeInfo(oldSub, curSub))
		delete(oldSubsUnusedSet, oldSub)
	}
	// now deleted
	for deletedSub, _ := range oldSubsUnusedSet {
		info.Subs = append(info.Subs, NewFolderChangeInfo(deletedSub, nil))
	}

	// same logic with files
	oldFilesMap := make(map[string]*internal.FileInfo)
	oldFilesUnusedSet := make(map[*internal.FileInfo]bool, len(old.Subs))
	for _, file := range old.Files {
		oldFilesUnusedSet[file] = true
		oldFilesMap[file.Name] = file
	}

	// all changed/identical/new are covered
	for _, curFile := range cur.Files {
		oldFile := oldFilesMap[curFile.Name]
		info.Files = append(info.Files, NewFileChangeInfo(oldFile, curFile))
		delete(oldFilesUnusedSet, oldFile)
	}
	// now deleted
	for deletedFile, _ := range oldFilesUnusedSet {
		info.Files = append(info.Files, NewFileChangeInfo(deletedFile, nil))
	}

	return &info
}

func Diff(old, cur *internal.FolderInfo) (*FolderChangeInfo, error) {
	err := checkInfos(old, cur)
	if err != nil {
		return nil, err
	}

	root := NewFolderChangeInfo(old, cur)
	root.searchChanges()

	return root, nil
}

func checkInfos(info1 *internal.FolderInfo, info2 *internal.FolderInfo) error {
	if info1.Data.Name != info2.Data.Name || info1.Data.Path != info2.Data.Path {
		return errors.New("root folders must be the same")
	}
	return nil
}
