package measure

import (
	"fedSharing/mainchain/configs"
	"fedSharing/mainchain/utils"
	"fmt"
	"time"
)

var MeasureStruct map[string][]map[string]int64

const (
	LOCALTRAIN = "local_train"
	AGGREGATE  = "server_aggregate"
	DIFF       = "diff_time"
	ASSESS     = "server_assess"
	SENDMODEL  = "send_model"
)

func DeferMeasureTime(Name string) func() {
	start := time.Now()
	return func() {
		configs.GlobalConfig.MeasureTimeViper.Set(Name, time.Since(start))
		err := configs.GlobalConfig.MeasureTimeViper.WriteConfigAs("./measure/out/time.json")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func MeasureTime(Name string, Who string, epoch int, start time.Time) {
	utils.ColorPrint(fmt.Sprintf("%s spend time %v\n", Name, time.Since(start)))
	MeasureStruct[Name][epoch][Who] = time.Since(start).Microseconds()
}

func WriteMeasureTimeToFile(path string) {
	configs.GlobalConfig.MeasureTimeViper.Set("fl_time_measure", MeasureStruct)
	err := configs.GlobalConfig.MeasureTimeViper.WriteConfigAs(path)
	if err != nil {
		fmt.Println(err)
	}
}
