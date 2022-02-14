package filemgr

type UserSpace struct {
	name   string
	path   string
	filedb string
}

var (
	root = "./data/user-space"
)

func SetSpaceRoot(r string) {
	root = r
}

// var (
// 	mUnit = map[string]float64{
// 		"B": 1.0 / (1024 * 1024),
// 		"K": 1.0 / 1024,
// 		"M": 1.0,
// 		"G": 1024,
// 		"T": 1024 * 1024,
// 	}
// )

// func (us UserSpace) String() string {
// 	return fmt.Sprintf("name: [%s]\npath: [%s]\ncapacity: [%d]M\nused: [%d]M\n",
// 		us.name, us.path, us.capacity, us.used)
// }

// func NewUserSpace(name string) *UserSpace {
// 	us := &UserSpace{
// 		name:     name,
// 		path:     filepath.Join(root, name),
// 		capacity: 30,
// 		used:     0,
// 	}
// 	if err := us.init(); err != nil {
// 		return nil
// 	}
// 	return us
// }

// func (us *UserSpace) init() error {
// 	if !fd.DirExists(us.path) {
// 		gio.MustCreateDir(us.path)
// 	}
// 	return nil
// }

// func (us *UserSpace) UpdateFields(path string, capacity int, updUsed bool) error {
// 	if path != "" {
// 		oldpath := us.path
// 		us.path = filepath.Join(root, path)
// 		if err := os.Rename(oldpath, us.path); err != nil {
// 			return err
// 		}
// 	}
// 	if capacity != 0 {
// 		sz, err := fd.DirSize(us.path, "m")
// 		if err != nil {
// 			return err
// 		}
// 		if capacity < int(sz) {
// 			return fmt.Errorf("new capacity size %dM is less than occupied size %dM", capacity, int(sz))
// 		}
// 		us.capacity = capacity
// 	}
// 	if updUsed {
// 		us.used = us.GetUsed("M")
// 	}
// 	return nil
// }

// func (us *UserSpace) Path() string {
// 	if fd.DirExists(us.path) {
// 		return us.path
// 	}
// 	return ""
// }

// // unit: B K M G T
// func (us *UserSpace) Capacity(unit string) int {
// 	val, ok := mUnit[unit]
// 	lk.FailOnErrWhen(!ok, "%v", fmt.Errorf("unsupported unit %v", unit))
// 	return int(math.Ceil(float64(us.capacity) / val))
// }

// func (us *UserSpace) GetUsed(unit string) int {
// 	sz, err := fd.DirSize(us.path, unit)
// 	lk.FailOnErr("%v", err)
// 	return int(math.Ceil(sz))
// }

// func (us *UserSpace) SaveFile(filename string, data []byte) error {
// 	ok := false
// 	defer func() { us.UpdateFields("", 0, ok) }()

// 	path := us.Path()
// 	if path == "" {
// 		return fmt.Errorf("no allocated space for [%s]", us.name)
// 	}

// 	dataSzM := int(math.Ceil(float64(len(data)) / (1024.0 * 1024.0)))
// 	if us.Capacity("M") < us.used+dataSzM {
// 		return fmt.Errorf("no available free space for [%s]", us.name)
// 	}

// 	oldpath := filepath.Join(path, filename)
// 	err := os.WriteFile(oldpath, data, os.ModePerm)
// 	if err != nil {
// 		return err
// 	}
// 	fType := GetFileType(oldpath)
// 	newpath := filepath.Join(us.Path(), fType, filename)
// 	gio.MustCreateDir(filepath.Dir(newpath))
// 	err = os.Rename(oldpath, newpath)
// 	if err == nil {
// 		ok = true
// 	}
// 	return err
// }
