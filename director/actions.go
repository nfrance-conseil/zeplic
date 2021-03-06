// Package director contains: actions.go - agent.go - director.go - slave.go
//
// Actions provides the functions to resolve the action to do
//
package director

import (
	"fmt"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/nfrance-conseil/zeplic/tools"
)

// Delete returns true if the snapshot should be deleted
func Delete(dataset string, DeleteList []string, prefix string, retention []string) (bool, []string) {
	var destroy bool
	var toDestroy []string

	D, W, M, Y := lib.Retention(retention)
	if D == 0 {
		D = 1
	}

	year, month, day := time.Now().Date()
	hour, min, _   := time.Now().Clock()
	loc, _ := time.LoadLocation("UTC")
	actual := time.Date(year, month, day, hour, min, 00, 0, loc)
	leap, monthDiff := tools.LengthMonth(year, month)

	// Remove snapshot with #NotWritten flag
	for f := 0; f < len(DeleteList); f++ {
		if strings.Contains(DeleteList[f], "#NotWritten") {
			DeleteList = append(DeleteList[:f], DeleteList[f+1:]...)
			f--
			continue
		}
	}

	// Remove snapshots with other prefix
	for g := 0; g < len(DeleteList); g++ {
		_, SnapshotName, _ := lib.InfoKV(DeleteList[g])
		RealPrefix := lib.Prefix(SnapshotName)
		if RealPrefix != prefix {
			DeleteList = append(DeleteList[:g], DeleteList[g+1:]...)
			continue
		}
	}

	// Sort the list of snapshots
	var newDeleteList []string
	for h := 0; h < len(DeleteList); h++ {
		name := tools.Reverse(DeleteList[h], ":")
		newDeleteList = append(newDeleteList, name)
	}
	newDeleteList = tools.Arrange(newDeleteList)

	// Keep the snapshot of reference
	for i := len(newDeleteList)-1; i > -1; i-- {
		if strings.Contains(newDeleteList[i], "#sent") {
			newDeleteList = append(newDeleteList[:i], newDeleteList[i+1:]...)
			break
		}
	}

	var today   []string
	var toweek  []string
	var tomonth []string
	var toyear  []string
	var noRange []string
	for j := len(newDeleteList)-1; j > -1; j-- {
		y, m, d, H, M, _ := lib.CreateTime(newDeleteList[j])
		expire := time.Date(y, m, d, H, M, 00, 0, loc)

		// Difference between last snapshot and the actual time
		diff := actual.Sub(expire).Hours()

		if diff < 24 {
			today = append(today, newDeleteList[j])
		} else if diff < 168 {
			toweek = append(toweek, newDeleteList[j])
		} else if diff < 672 && monthDiff == "notleap" {
			tomonth = append(tomonth, newDeleteList[j])
		} else if diff < 696 && monthDiff == "leap" {
			tomonth = append(tomonth, newDeleteList[j])
		} else if diff < 720 && monthDiff == "short" {
			tomonth = append(tomonth, newDeleteList[j])
		} else if diff < 744 && monthDiff == "long" {
			tomonth = append(tomonth, newDeleteList[j])
		} else if leap == false && diff < 8760 || leap == true && diff < 8784 {
			toyear = append(toyear, newDeleteList[j])
		} else {
			noRange = append(noRange, newDeleteList[j])
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
					if k == len(toweek)-2 {
						queue = append(queue, toweek[k+1])
					}
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
					if k == len(toyear)-2 {
						queue = append(queue, toyear[k+1])
					}
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
			for p := 0; p < len(DeleteList); p++ {
				if strings.Contains(DeleteList[p], queue[n]) {
					toDestroy = append(toDestroy, DeleteList[p])
				}
			}
		}
	}
	return destroy, toDestroy
}

// NewSnapshot returns true if a new snapshot should be created and its name
func NewSnapshot(TakeList []string, cron string, prefix string) (bool, string) {
	var take bool
	var SnapshotName string

	// Actual time
	year, month, day := time.Now().Date()
	hour, min, _ := time.Now().Clock()
	loc, _ := time.LoadLocation("UTC")
	actual := time.Date(year, month, day, hour, min, 00, 0, loc)

	// Struct for cron
	cMinute, cHour, cMonthday, cMonth, cWeekday := tools.Crontab(cron)

	// Remove KV pair for sync
	var SnapSync string
	if len(TakeList) > 1 {
		for g := 0; g < len(TakeList); g++ {
			if strings.Contains(TakeList[g], "zCHECK") {
				SnapSync = TakeList[g]
				TakeList = append(TakeList[:g], TakeList[g+1:]...)
				g--
			}
		}

		// Remove snapshots with other prefix
		for h := 0; h < len(TakeList); h++ {
			_, SnapshotName, _ := lib.InfoKV(TakeList[h])
			RealPrefix := lib.Prefix(SnapshotName)
			if RealPrefix != prefix {
				TakeList = append(TakeList[:h], TakeList[h+1:]...)
				continue
			}
		}
	}

	// Sort the list of snapshtos
	for i := 0; i < len(TakeList); i++ {
		_, name, _ := lib.InfoKV(TakeList[i])
		TakeList[i] = name
	}
	TakeList = tools.Arrange(TakeList)

	// Last snapshot time
	if len(TakeList) == 0 {
		TakeList = append(TakeList, SnapSync)
	}
	LastSnapshot := TakeList[len(TakeList)-1]
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

							if diff2 <= diff1 && diff2 > 0 && diff1 > diff3 && diff3 < 21600 {
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

								if diff2 <= diff1 && diff2 > 0 && diff1 > diff3 && diff3 < 21600 {
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
func Send(dataset string, SentList []string, SyncPolicy string, prefix string) (bool, string) {
	var send bool
	var SnapshotUUID string

	// Actual time
	year, month, day := time.Now().Date()
	hour, min, _ := time.Now().Clock()
	loc, _ := time.LoadLocation("UTC")
	actual := time.Date(year, month, day, hour, min, 00, 0, loc)

	// Struct for cron
	cMinute, cHour, cMonthday, cMonth, cWeekday := tools.Crontab(SyncPolicy)

	// Remove all snapshots that contains the flag #NotWritten or #sent
	for f := 0; f < len(SentList); f++ {
		if strings.Contains(SentList[f], "#NotWritten") || strings.Contains(SentList[f], "#sent") {
			SentList = append(SentList[:f], SentList[f+1:]...)
			f--
		}
	}

	// Remove snapshots with other prefix
	for g := 0; g < len(SentList); g++ {
		_, SnapshotName, _ := lib.InfoKV(SentList[g])
		RealPrefix := lib.Prefix(SnapshotName)
		if RealPrefix != prefix {
			SentList = append(SentList[:g], SentList[g+1:]...)
			continue
		}
	}

	// Sort the list of snapshtos
	var newSentList []string
	for h := 0; h < len(SentList); h++ {
		_, name, _ := lib.InfoKV(SentList[h])
		newSentList = append(newSentList, name)
	}
	newSentList = tools.Arrange(newSentList)

	// Take the snapshots with the same prefix
	var LastSnapshot string
	if len(newSentList) > 0 {
		for i := len(newSentList)-1; i > -1; i-- {
			LastSnapshot = newSentList[i]
			// Checking the syncrhonization policy
			if SyncPolicy == "asap" {
				send = true
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
		}
	}

	// Extract the uuid of snapshot
	for z := 0; z < len(SentList); z++ {
		LastSnapshot := tools.NumberMonthReverse(LastSnapshot)
		if strings.Contains(SentList[z], LastSnapshot) {
			SnapshotUUID = tools.Before(SentList[z], ":")
			break
		} else {
			continue
		}
	}
	return send, SnapshotUUID
}
