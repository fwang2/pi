package cmd

// find -type 			find either file or directory
// find -empty 			find empty file or directory
// find -exec rm 		process found file
// find -name xyz		find xyz, exactly
// find -name "*.txt"   	find file or directory that matches the name
// find -exec chmod 777 	process found file
// find /path			find all files under path
// find / -size 50M-100M 	find size in a range
// find / -size +50M		find size bigger
// find / -size -100M   	find size smaller
// find / -size +1G -exec rm  	find and delete
// find / -mtime 50		find files modified 50 days back
// find / -atime 50 		find files are accessed 50 days back
// find / -mtime 50-100		find modified in between 50 to 100 days back
// find / -cmin -60		find changed file in last 1 hour
// find / -mmin -60		find modified file in last 1 hour
// find / -amin -60		find access files in last 1 hour
// find / -exec save		save found

import (
	"regexp"
	"time"

	"github.com/fwang2/pi/fs"
	"github.com/fwang2/pi/util"
	"github.com/spf13/cobra"
)

//var log = util.NewLogger()

var findc = &fs.FindControl{}

var fname string
var fsize string
var ftype string
var apparent bool
var delete bool
var atime string
var ctime string
var mtime string

var f_map = map[string]bool{
	fs.F_FILE: true, // file
	fs.F_DIR:  true, // directory
	//fs.F_SOCKET:  true, // socket
	fs.F_SYMLINK: true, // sym link
	//fs.F_CHAR:    true, // char special
	//fs.F_PIPE:    true, // pipe
	//fs.F_BLOCK:   true, // block special
}

func init() {
	findCmd.Flags().StringVar(&fname, "name", "", "On name pattern")
	findCmd.Flags().StringVar(&fsize, "size", "", "On file size")
	findCmd.Flags().BoolVar(&apparent, "apparent", false, "Use apparent size")
	findCmd.Flags().StringVar(&ftype, "type", "", "On file type")
	findCmd.Flags().StringVar(&atime, "atime", "", "Access time (e.g 4h30m)")
	findCmd.Flags().StringVar(&ctime, "ctime", "", "Creation time (e.g 4h30m)")
	findCmd.Flags().StringVar(&mtime, "mtime", "", "Modification time (e.g 4h30m)")
	findCmd.Flags().BoolVar(&delete, "delete", false, "delete files")

	rootCmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "A subset of Unix find-alike functions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// validate flags
		if fname != "" {
			findc.Name = fname
			findc.Flags = fs.Set(findc.Flags, fs.FB_NAME)
		}
		if fsize != "" {
			parse_fsize(fsize)
			findc.Flags = fs.Set(findc.Flags, fs.FB_SIZE)
			findc.Apparent = apparent
		}

		if delete {
			findc.DeleteFlag = true
		}

		if ftype != "" {
			parse_ftype(ftype)
		} else {
			// if not ftype specified, default it f
			findc.Flags = fs.Set(findc.Flags, fs.FB_TYPE_F)
			//findc.Flags = fs.Set(findc.Flags, fs.FB_TYPE_D)
			//findc.Flags = fs.Set(findc.Flags, fs.FB_TYPE_L)
		}
		parse_time()

		log.Debugf("findc.Flags = %b", findc.Flags)
		// Determine path
		var ws *fs.WalkStat = new(fs.WalkStat)
		ws.NumOfWorkers = NumOfWorkers
		ws.RootPath = fs.ParseRootPath(args)
		ws.NumOfWorkers = NumOfWorkers
		var wc *fs.WalkControl = new(fs.WalkControl)
		wc.DoProgress = false
		wc.Findc = findc
		fs.RunProfile(wc, ws)
	},
}

func parse_time() {
	// for now, we only handle time util now.
	var t string
	if atime != "" {
		t = atime
		findc.Flags = fs.Set(findc.Flags, fs.FB_ATIME)
	} else if mtime != "" {
		t = mtime
		findc.Flags = fs.Set(findc.Flags, fs.FB_MTIME)
	} else if ctime != "" {
		t = ctime
		findc.Flags = fs.Set(findc.Flags, fs.FB_CTIME)
	} else {
		return
	}

	du, neg, err := util.ParseDuration(t)
	if err != nil {
		log.Fatalf("Wrong time format: %s", t)
	}
	if neg {
		findc.EndTime = time.Now()
		findc.StartTime = findc.EndTime.Add(-du)
	} else {
		findc.StartTime = time.Unix(0, 0)
		findc.EndTime = time.Now().Add(-du)
	}

	log.Debugf("start_time = %v, end_time = %v", findc.StartTime, findc.EndTime)

}
func parse_ftype(s string) {
	if !f_map[s] {
		log.Fatalf("can't parse file type: %s. Must be one of {dfl}", s)
	}
	switch s {
	case fs.F_DIR:
		findc.Flags = fs.Set(findc.Flags, fs.FB_TYPE_D)
	case fs.F_FILE:
		findc.Flags = fs.Set(findc.Flags, fs.FB_TYPE_F)
	case fs.F_SYMLINK:
		findc.Flags = fs.Set(findc.Flags, fs.FB_TYPE_L)
	}
}

func parse_fsize(s string) {
	// 4k, -4k, +4k are valid
	// re := regexp.MustCompile(`(?P<x>\+|\-)?(?P<y>[[:digit:]]+)(?P<z>(c|C|k|K|m|M|g|G|t|T)?)`)
	re := regexp.MustCompile(`(\+|\-)?([[:digit:]]+)(c|C|k|K|m|M|g|G|t|T)?`)
	out := re.FindStringSubmatch(s)
	if len(out) == 0 {
		log.Fatalf("Can't parse --size: %s", fsize)
	}

	// DEBU[0000] -size: [+400k + 400 k], length=4
	// DEBU[0000] 	 k=0, v=+400k
	// DEBU[0000] 	 k=1, v=+
	// DEBU[0000] 	 k=2, v=400
	// DEBU[0000] 	 k=3, v=k

	log.Debugf("-size: %v, length=%d", out, len(out))

	if out[1] == "+" {
		findc.SizeOp = fs.GREAT_THAN
	} else if out[1] == "-" {
		findc.SizeOp = fs.LESS_THAN
	} else {
		findc.SizeOp = fs.EQUAL
	}

	// no unit is given, default to c
	if out[3] == "" {
		out[3] = "c"
	}

	findc.Size = util.StrBytes(out[2] + out[3])
	findc.Flags = fs.Set(findc.Flags, fs.FB_SIZE)
	log.Debugf("fsize threshold = %d\n", findc.Size)
}
