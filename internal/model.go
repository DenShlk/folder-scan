package internal

type FSData struct {
	Name   string      `json:"name,omitempty"`
	Path   string      `json:"path,omitempty"`
	Size   int64       `json:"size,omitempty"`
	Parent *FolderInfo `json:"-"`
}

type FolderInfo struct {
	Data *FSData `json:"data"`

	FilesSize      int64 `json:"filesSize,omitempty"`
	SubfoldersSize int64 `json:"subfoldersSize,omitempty"`

	Subs  []*FolderInfo `json:"subs,omitempty"`
	Files []*FileInfo   `json:"files,omitempty"`
}

func (i *FolderInfo) CalcSize() int64 {
	i.SubfoldersSize = 0
	for _, folder := range i.Subs {
		i.SubfoldersSize += folder.CalcSize()
	}
	i.Data.Size = i.SubfoldersSize + i.FilesSize
	return i.Data.Size
}

type FileInfo FSData
