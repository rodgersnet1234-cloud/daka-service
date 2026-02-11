package config

import "daka/model"

var DailyActivity = map[string]int{
	"1星": 100,
	"2星": 1000,
	"3星": 5000,
	"4星": 20000,
	"5星": 120000,
	"6星": 260000,
}

func ReturnActivity() []model.LevelThreshold {

	data := []model.LevelThreshold{
		{"1星", 100},
		{"2星", 1000},
		{"3星", 5000},
		{"4星", 20000},
		{"5星", 120000},
		{"6星", 260000},
	}
	return data

}

var PackageValue = map[int]int{
	100:    1,
	1000:   2,
	5000:   3,
	20000:  4,
	120000: 5,
	260000: 6,
}

var PackageLevel = map[string]string{
	"10":    "一星店铺",
	"100":   "二星体验店铺",
	"500":   "三星体验店铺",
	"1000":  "四星体验店铺",
	"3000":  "五星体验店铺",
	"5000":  "六星体验店铺",
	"10000": "七星体验店铺",
	"30000": "形象店铺",
	"50000": "旗舰店铺",
}

var ShortPackageLevel = map[string]string{
	"10":    "一星店铺",
	"100":   "二星店铺",
	"500":   "三星店铺",
	"1000":  "四星店铺",
	"3000":  "五星店铺",
	"5000":  "六星店铺",
	"10000": "七星店铺",
	"30000": "形象店铺",
	"50000": "旗舰店铺",
}

var DailyPackage = map[int]float64{
	10:    0.0253,
	100:   0.2600,
	500:   1.3000,
	1000:  2.6666,
	3000:  8.2000,
	5000:  14.0000,
	10000: 28.6666,
	30000: 87.0000,
	50000: 146.6666,
}

var PackageOrder = map[int]int{
	10:    1,
	100:   2,
	500:   3,
	1000:  4,
	5000:  5,
	10000: 6,
}

var PackageInfo = map[string]int{
	"1星": 10,
	"2星": 100,
	"3星": 500,
	"4星": 1000,
	"5星": 5000,
	"6星": 10000,
}

var DailyHalf = map[string]int{
	"1星": 50,
	"2星": 500,
	"3星": 2500,
	"4星": 10000,
	"5星": 50000,
	"6星": 130000,
}

var DailyProfit = map[int]float64{
	1: 0.3,
	2: 0.4,
	3: 0.5,
	4: 0.6,
	5: 0.7,
	6: 0.8,
}

var SystemSplit = map[int]float64{
	2: 0.15,
	3: 0.12,
	4: 0.08,
	5: 0.06,
	6: 0.04,
}

var ErrorMsgs = map[string]int{
	"UpdateUserShouyi":          1,
	"UpdateUserJifen":           2,
	"UpdatePackageLogDay":       3,
	"UpdateUserAV":              4,
	"UpdateUserInviterEarnings": 5,
}

// var DailyPackage = map[int]float64{
// 	10:    0.076,
// 	100:   0.078,
// 	500:   0.08,
// 	1000:  0.08,
// 	3000:  0.082,
// 	5000:  0.084,
// 	10000: 0.086,
// 	30000: 0.087,
// 	50000: 0.088,
// }

// var DailyPackage = map[int]float64{
// 	10:    0.2533,
// 	100:   0.26,
// 	500:   0.2667,
// 	1000:  0.2667,
// 	3000:  0.2733,
// 	5000:  0.28,
// 	10000: 0.2867,
// 	30000: 0.29,
// 	50000: 0.2933,
// }
