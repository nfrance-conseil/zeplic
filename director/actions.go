// Package director contains: actions.go - agent.go - director.go - slave.go
//
// Actions provides the functions to resolve the action to do
//
package director

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/nfrance-conseil/zeplic/tools"
)

// Arrange sorts the list of snapshots
func Arrange(SnapshotsList []string) []string {
	for i := 0; i < len(SnapshotsList); i++ {
		SnapshotsList[i] = tools.NumberMonth(SnapshotsList[i])
	}
	sort.Strings(SnapshotsList)
	return SnapshotsList
}

// Delete returns true if the snapshot should be deleted
func Delete(dataset string, SnapshotsList []string, prefix string, retention string) (bool, []string) {
	var destroy bool
	var toDestroy []string

	D, W, M, Y := tools.Retention(retention)
	if D == 0 {
		D = 1
	}

	year, month, day := time.Now().Date()
	hour, min, _   := time.Now().Clock()
	loc, _ := time.LoadLocation("UTC")
	actual := time.Date(year, month, day, hour, min, 00, 0, loc)

	leap, monthDiff := tools.LengthMonth(year, month)

	var newList []string
	for f := 0; f < len(SnapshotsList); f++ {
		snapName := tools.Reverse(SnapshotsList[f], ":")
		newList = append(newList, snapName)
	}

	// Remove all snapshots of others datasets
	for g := 0; g < len(newList); g++ {
		RealDataset := lib.DatasetName(newList[g])
		if RealDataset != dataset {
			newList = append(newList[:g], newList[g+1:]...)
			g--
			continue
		} else {
			continue
		}
	}

	// Remove all snapshots with others prefix
	for h := 0; h < len(newList); h++ {
		RealPrefix := lib.Prefix(newList[h])
		if RealPrefix != prefix {
			newList = append(newList[:h], newList[h+1:]...)
			h--
			continue
		} else {
			continue
		}
	}

	// Sort the list of snapshots
	newList = Arrange(newList)

	// Keep the snapshot of reference
	for i := len(newList)-1; i > -1; i-- {
		if strings.Contains(newList[i], "#sent") {
			newList = append(newList[:i], newList[i+1:]...)
			i++
			break
		}
	}

	var today   []string
	var toweek  []string
	var tomonth []string
	var toyear  []string
	var noRange []string
	for j := len(newList)-1; j > -1; j-- {
		y, m, d, H, M, _ := lib.CreateTime(newList[j])
		expire := time.Date(y, m, d, H, M, 00, 0, loc)

		// Difference between last snapshot and the actual time
		diff := actual.Sub(expire).Hours()

		if diff < 24 {
			today = append(today, newList[j])
		} else if diff < 168 {
			toweek = append(toweek, newList[j])
		} else if diff < 672 && monthDiff == "notleap" {
			tomonth = append(tomonth, newList[j])
		} else if diff < 696 && monthDiff == "leap" {
			tomonth = append(tomonth, newList[j])
		} else if diff < 720 && monthDiff == "short" {
			tomonth = append(tomonth, newList[j])
		} else if diff < 744 && monthDiff == "long" {
			tomonth = append(tomonth, newList[j])
		} else if leap == false && diff < 8760 || leap == true && diff < 8784 {
			toyear = append(toyear, newList[j])
		} else {
			noRange = append(noRange, newList[j])
		}
	}

	var queue []string
	// No range snapshots
	if len(noRange) > 0 {
		for k := 0; k < len(noRange); k++ {
			queue = append(queue, noRange[k])
		}
	}

	// Today snapshots out of daily range
	if len(today) > D {
		for k := len(today)-1; k > D-1; k-- {
			queue = append(queue, today[k])
		}
	}

	// Toweek snapshots out of week range
	if W > 0 {
		var checkDay[]string
		// Comparison between pairs of snapshots
		for k := 0; k < len(toweek)-1; k++ {
			_, _, d1, _, _, _ := lib.CreateTime(toweek[k])
			_, _, d2, _, _, _ := lib.CreateTime(toweek[k+1])
			if d1 == d2 {
				checkDay = append(checkDay, toweek[k])
				if len(checkDay) > W {
					queue = append(queue, checkDay[len(checkDay)-1])
					checkDay = append(checkDay[:len(checkDay)-1])
				} else {
					continue
				}
			} else {
				checkDay = append(checkDay, toweek[k])
				if len(checkDay) > W {
					queue = append(queue, checkDay[len(checkDay)-1])
					checkDay = append(checkDay[:len(checkDay)-1])
				}
				checkDay = append(checkDay[len(checkDay):])
				continue
			}
		}
	} else {
		for k := 0; k < len(toweek); k++ {
			queue = append(queue, toweek[k])
		}
	}

	// Tomonth snapshots out of month range
	if M > 0 {
		var checkWeek1 []string
		var checkWeek2 []string
		var checkWeek3 []string
		var checkWeek4 []string
		// Comparison between pairs of snapshots
		for k := 0; k < len(tomonth); k++ {
			y, m, d, HH, MM, _ := lib.CreateTime(tomonth[k])
			snap := time.Date(y, m, d, HH, MM, 00, 0, loc)
			diff := actual.Sub(snap).Hours()
			if diff < 336 {
				checkWeek1 = append(checkWeek1, tomonth[k])
				if len(checkWeek1) > M {
					queue = append(queue, checkWeek1[len(checkWeek1)-1])
					checkWeek1 = append(checkWeek1[:len(checkWeek1)-1])
				} else {
					continue
				}
			} else if diff >= 336 && diff < 504 {
				checkWeek2 = append(checkWeek2, tomonth[k])
				if len(checkWeek2) > M {
					queue = append(queue, checkWeek2[len(checkWeek2)-1])
					checkWeek2 = append(checkWeek2[:len(checkWeek2)-1])
				} else {
					continue
				}
			} else if diff >= 504 && diff < 672 {
				checkWeek3 = append(checkWeek3, tomonth[k])
				if len(checkWeek3) > M {
					queue = append(queue, checkWeek3[len(checkWeek3)-1])
					checkWeek3 = append(checkWeek3[:len(checkWeek3)-1])
				} else {
					continue
				}
			} else {
				checkWeek4 = append(checkWeek4, tomonth[k])
				if len(checkWeek4) > M {
					queue = append(queue, checkWeek4[len(checkWeek4)-1])
					checkWeek4 = append(checkWeek4[:len(checkWeek4)-1])
				} else {
					continue
				}
			}
		}
	} else {
		for k := 0; k < len(tomonth); k++ {
			queue = append(queue, tomonth[k])
		}
	}

	// Toyear snapshots out of year range
	if Y > 0 {
		var checkMonth []string
		// Comparison between pairs of snapshots
		for k := 0; k < len(toyear)-1; k++ {
			_, m1, _, _, _, _ := lib.CreateTime(toyear[k])
			_, m2, _, _, _, _ := lib.CreateTime(toyear[k+1])
			if m1 == m2 {
				checkMonth = append(checkMonth, toyear[k])
				if len(checkMonth) > Y {
					queue = append(queue, checkMonth[len(checkMonth)-1])
					checkMonth = append(checkMonth[:len(checkMonth)-1])
				} else {
					continue
				}
			} else {
				checkMonth = append(checkMonth, toyear[k])
				if len(checkMonth) > Y {
					queue = append(queue, checkMonth[len(checkMonth)-1])
					checkMonth = append(checkMonth[:len(checkMonth)-1])
				}
				checkMonth = append(checkMonth[len(checkMonth):])
				continue
			}
		}
	} else {
		for k := 0; k < len(toyear); k++ {
			queue = append(queue, toyear[k])
		}
	}

	// Should I destroy snapshots?
	if len(queue) > 0 {
		destroy = true
		for m := 0; m < len(queue); m++ {
			queue[m] = tools.NumberMonthReverse(queue[m])
		}
		for n := 0; n < len(queue); n++ {
			for p := 0; p < len(SnapshotsList); p++ {
				if strings.Contains(SnapshotsList[p], queue[n]) {
					toDestroy = append(toDestroy, SnapshotsList[p])
				}
			}
		}
	}
	return destroy, toDestroy
}

// NewSnapshot returns true if a new snapshot should be created and its name
func NewSnapshot(SnapshotsList []string, cron string, prefix string) (bool, string) {
	var take bool
	var SnapshotName string

	// Actual time
	year, month, day := time.Now().Date()
	hour, min, _ := time.Now().Clock()
	loc, _ := time.LoadLocation("UTC")
	actual := time.Date(year, month, day, hour, min, 00, 0, loc)

	// Struct for cron
	cMinute, cHour, cMonthday, cMonth, cWeekday := tools.Crontab(cron)

	// Sort the list of snapshtos
	var list []string
	for h := 0; h < len(SnapshotsList); h++ {
		_, name, _ := lib.InfoKV(SnapshotsList[h])
		list = append(list, name)
	}
	list = Arrange(list)

	// Take the snapshots with the same prefix
	var LastSnapshot string
	for i := len(list)-1; i > -1; i-- {
		LastSnapshot = list[i]
		SnapPrefix := lib.Prefix(LastSnapshot)
		if SnapPrefix == prefix || SnapPrefix == "zCHECK" {
			break
		} else {
			continue
		}
	}

	// Last snapshot time
	y, m, d, H, M, _ := lib.CreateTime(LastSnapshot)
	last := time.Date(y, m, d, H, M, 00, 0, loc)
	diff1 := actual.Sub(last).Seconds()

	// Comparison with cron format
	if year == y {
		for j := len(cMonth)-1; j > -1; j-- {
			for k := len(cMonthday)-1; k > -1; k-- {
				for m := len(cHour)-1; m > -1; m-- {
					for n := len(cMinute)-1; n > -1; n-- {
						if take == false {
							inter := time.Date(year, cMonth[j], cMonthday[k], cHour[m], cMinute[n], 00, 0, loc)
							wdayInter := inter.Weekday()
							diff2 := inter.Sub(last).Seconds()
							diff3 := actual.Sub(inter).Seconds()

							if diff2 <= diff1 && diff2 > 0 && diff1 > diff3 {
								// Weekday
								for p := 0; p < len(cWeekday); p++ {
									if cWeekday[p] == wdayInter {
										take = true
										yyyy, mm, dd := inter.Date()
										HH, MM, _ := inter.Clock()
										SnapshotName = fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:00", prefix, yyyy, mm, dd, HH, MM)
										break
									} else {
										continue
									}
								}
							} else if diff1 <= diff3 {
								break
							} else {
								continue
							}
						}
					}
				}
			}
		}
	} else {
		for r := year; r > year-2; r-- {
			for j := len(cMonth)-1; j > -1 ; j-- {
				for k := len(cMonthday)-1; k > -1; k-- {
					for m := len(cHour)-1; m > -1; m-- {
						for n := len(cMinute)-1; n > -1; n-- {
							if take == false {
								inter := time.Date(r, cMonth[j], cMonthday[k], cHour[m], cMinute[n], 00, 0, loc)
								wdayInter := inter.Weekday()
								diff2 := inter.Sub(last).Seconds()
								diff3 := actual.Sub(inter).Seconds()

								if diff2 <= diff1 && diff2 > 0 && diff1 > diff3 {
									// Weekday
									for p := 0; p < len(cWeekday); p++ {
										if cWeekday[p] == wdayInter {
											take = true
											yyyy, mm, dd := inter.Date()
											HH, MM, _ := inter.Clock()
											SnapshotName = fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:00", prefix, yyyy, mm, dd, HH, MM)
											break
										} else {
											continue
										}
									}
								} else if diff1 <= diff3 {
									break
								} else {
									continue
								}
							}
						}
					}
				}
			}
		}
	}
	return take, SnapshotName
}

// Send returns true if the snapshot should be sent
func Send(dataset string, SnapshotsList []string, SyncPolicy string, prefix string) (bool, string) {
	var send bool
	var SnapshotUUID string

	// Remove all snapshots that contains the flag #sent
	for f := 0; f < len(SnapshotsList); f++ {
		if strings.Contains(SnapshotsList[f], "#sent") {
			SnapshotsList = append(SnapshotsList[:f], SnapshotsList[f+1:]...)
			f--
		}
	}

	// Actual time
	year, month, day := time.Now().Date()
	hour, min, _ := time.Now().Clock()
	loc, _ := time.LoadLocation("UTC")
	actual := time.Date(year, month, day, hour, min, 00, 0, loc)

	// Struct for cron
	cMinute, cHour, cMonthday, cMonth, cWeekday := tools.Crontab(SyncPolicy)

	// Sort the list of snapshtos
	var list []string
	for h := 0; h < len(SnapshotsList); h++ {
		_, name, _ := lib.InfoKV(SnapshotsList[h])
		list = append(list, name)
	}
	list = Arrange(list)

	// Take the snapshots with the same prefix
	var LastSnapshot string
	for i := len(list)-1; i > -1; i-- {
		LastSnapshot = list[i]
		SnapPrefix := lib.Prefix(LastSnapshot)
		if SnapPrefix == prefix {
			// Checking the syncrhonization policy
			if SyncPolicy == "asap" {
				send= true
				break
			} else {
				// Last snapshot time
				y, m, d, H, M, _ := lib.CreateTime(LastSnapshot)
				last := time.Date(y, m, d, H, M, 00, 0, loc)
				diff1 := actual.Sub(last).Seconds()

				// Comparison with cron format
				if year == y {
					for j := 0; j < len(cMonth); j++ {
						for k := 0; k < len(cMonthday); k++ {
							for m := 0; m < len(cHour); m++ {
								for n := 0; n < len(cMinute); n++ {
									if send == false {
										inter := time.Date(year, cMonth[j], cMonthday[k], cHour[m], cMinute[n], 00, 0, loc)
										wdayInter := inter.Weekday()
										diff2 := inter.Sub(last).Seconds()

										if diff2 > 0 && diff2 < diff1 {
											// Weekday
											for p := 0; p < len(cWeekday); p++ {
												if cWeekday[p] == wdayInter {
													send = true
													break
												} else {
													continue
												}
											}
										} else {
											continue
										}
									}
								}
							}
						}
					}
				} else {
					for r := y; r < year+1; r++ {
						for j := 0; j < len(cMonth); j++ {
							for k := 0; k < len(cMonthday); k++ {
								for m := 0; m < len(cHour); m++ {
									for n := 0; n < len(cMinute); n++ {
										if send == false {
											inter := time.Date(r, cMonth[j], cMonthday[k], cHour[m], cMinute[n], 00, 0, loc)
											wdayInter := inter.Weekday()
											diff2 := inter.Sub(last).Seconds()

											if diff2 > 0 && diff2 < diff1 {
												// Weekday
												for p := 0; p < len(cWeekday); p++ {
													if cWeekday[p] == wdayInter {
														send = true
														break
													} else {
														continue
													}
												}
											} else {
												continue
											}
										}
									}
								}
							}
						}
					}
				}
			}
		} else {
			continue
		}
	}

	// Extract the uuid of snapshot
	for z := 0; z < len(SnapshotsList); z++ {
		LastSnapshot := tools.NumberMonthReverse(LastSnapshot)
		if strings.Contains(SnapshotsList[z], LastSnapshot) {
			SnapshotUUID = tools.Before(SnapshotsList[z], ":")
			break
		} else {
			continue
		}
	}
	return send, SnapshotUUID
}
