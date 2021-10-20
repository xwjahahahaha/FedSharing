package measure

import (
	"fedSharing/mainchain/configs"
	"fedSharing/mainchain/utils"
	"fmt"
	"time"
)

func DeferMeasureTime(Name string) func()  {
	start := time.Now()
	return func() {
		configs.GlobalConfig.MeasureTimeViper.Set(Name, time.Since(start))
		err := configs.GlobalConfig.MeasureTimeViper.WriteConfigAs("./measure/out/time.json")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func MeasureTime(Name string, start time.Time)  {
	utils.ColorPrint(fmt.Sprintf("%s spend time %v\n", Name, time.Since(start)))
	configs.GlobalConfig.MeasureTimeViper.Set(Name, fmt.Sprintf("%v", time.Since(start)))
}

func WriteMeasureTimeToFile(path string){
	err := configs.GlobalConfig.MeasureTimeViper.WriteConfigAs(path)
	if err != nil {
		fmt.Println(err)
	}
}