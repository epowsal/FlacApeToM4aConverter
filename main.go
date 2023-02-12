// FlacApeToM4aConverter project main.go
package main

import (
	"charset"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"linbo.ga/toolfunc"
)

var programdir string

//config value
var rate string = "256K"

func CopyFile(srcName, dstName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

func StdFormatCueFile(path string) {
	cuedata, err := ioutil.ReadFile(path)
	if err == nil {
		cuedata2 := string(cuedata)
		cuedatals := strings.Split(cuedata2, "\n")
		binhead := true
		stdname := path[strings.LastIndex(path, "\\")+1:]
		stdname = stdname[0:strings.LastIndex(stdname, ".")]
		for ind, _ := range cuedatals {
			cuedatals[ind] = strings.Trim(cuedatals[ind], "\r\n\t ")
			if len(cuedatals[ind]) >= 5 && cuedatals[ind][0:5] != "TRACK" && binhead {
				//fmt.Println("cuedatals[ind][0:4]:", cuedatals[ind][0:4])
				if cuedatals[ind][0:4] == "FILE" && strings.LastIndex(cuedatals[ind], ".") != -1 {
					cuedatals[ind] = cuedatals[ind][0:strings.Index(cuedatals[ind], "\"")+1] + stdname + cuedatals[ind][strings.LastIndex(cuedatals[ind], "."):]
				}
			} else {
				binhead = false
				if len(cuedatals[ind]) >= 5 && cuedatals[ind][0:5] == "TRACK" {
					cuedatals[ind] = "  " + cuedatals[ind]
				} else {
					cuedatals[ind] = "    " + cuedatals[ind]
				}
			}
		}
		stdcuectt := ""
		for ind, _ := range cuedatals {
			if cuedatals[ind] != "" {
				stdcuectt += cuedatals[ind] + "\n"
			}
		}
		ioutil.WriteFile(path, []byte(stdcuectt), 0666)
	}
}

func ScanDirForStd(collectiondir string) bool {
	//fmt.Println("collectiondir:", collectiondir)
	if collectiondir[len(collectiondir)-1:] != "\\" {
		collectiondir += "\\"
	}
	allname, err := ioutil.ReadDir(collectiondir)
	if err == nil {
		//do find one
		for i := 0; i < len(allname); i++ {
			if allname[i].Name() == "FlacApeExtractTemp" {
				continue
			}
			fname := collectiondir + allname[i].Name()
			//fmt.Println("fname:", fname)
			if allname[i].IsDir() {
				bhavezip := false
				aels := []string{"(ED2000.COM).rar", "(ED2000.COM).zip", "(ED2000.COM).7z", ".rar", ".zip", ".7z", ".(ED2000.COM).rar", ".(ED2000.COM).zip", ".(ED2000.COM).7z"}
				for _, ext := range aels {
					dirst, err := os.Stat(collectiondir + allname[i].Name() + ext)
					if err == nil && !dirst.IsDir() {
						bhavezip = true
						DeleteDir(collectiondir + allname[i].Name())
						break
					}
				}
				if bhavezip == false {
					ScanDirForStd(collectiondir + allname[i].Name())
				}
			} else if strings.LastIndex(fname, ".") != -1 {
				extstr := strings.ToLower(fname[strings.LastIndex(fname, "."):])
				//fmt.Println(extstr)
				if extstr == ".cue" {
					StdFormatCueFile(fname)
				}
			}
		}
	}
	return false
}

var cpuusemap sync.Map

func cvtthread1(cpuid, startpos, inputpath, lenparam, lenstr, tracktitle, trackperformer, filename, cuetitle, cuedate, tracktag, outm4apath string) {
	fmt.Println("convert to:", outm4apath)
	fmt.Println("parm1:", programdir+"ffmpeg\\ffmpeg.exe", "-y", "-ss", startpos, "-i", inputpath, lenparam, lenstr, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-codec:a:0", "libfaac", "-metadata", "title="+tracktitle, "-metadata", "artist="+trackperformer, "-metadata", "comment="+filename, "-metadata", "album="+cuetitle, "-metadata", "date="+cuedate, "-metadata", "track="+tracktag, "-movflags", "faststart", outm4apath)
	cmd := exec.Command(programdir+"ffmpeg\\ffmpeg.exe", "-y", "-ss", startpos, "-i", inputpath, lenparam, lenstr, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-codec:a:0", "libfaac", "-metadata", "title="+tracktitle, "-metadata", "artist="+trackperformer, "-metadata", "comment="+filename, "-metadata", "album="+cuetitle, "-metadata", "date="+cuedate, "-metadata", "track="+tracktag, "-movflags", "faststart", outm4apath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	cpuusemap.Store(cpuid, "0")
}

func cvtthread_lasttrack(cpuid, startpos, inputpath, tracktitle, trackperformer, filename, cuetitle, cuedate, tracktag, outm4apath string) {
	fmt.Println("convert to:", outm4apath)
	fmt.Println("parm1:", programdir+"ffmpeg\\ffmpeg.exe", "-y", "-ss", startpos, "-i", inputpath, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-codec:a:0", "libfaac", "-metadata", "title="+tracktitle, "-metadata", "artist="+trackperformer, "-metadata", "comment="+filename, "-metadata", "album="+cuetitle, "-metadata", "date="+cuedate, "-metadata", "track="+tracktag, "-movflags", "faststart", outm4apath)
	cmd := exec.Command(programdir+"ffmpeg\\ffmpeg.exe", "-y", "-ss", startpos, "-i", inputpath, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-codec:a:0", "libfaac", "-metadata", "title="+tracktitle, "-metadata", "artist="+trackperformer, "-metadata", "comment="+filename, "-metadata", "album="+cuetitle, "-metadata", "date="+cuedate, "-metadata", "track="+tracktag, "-movflags", "faststart", outm4apath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	cpuusemap.Store(cpuid, "0")
}

func cvtthread2(cpuid, fpath, outm4apath string) {
	fmt.Println("convert to:", outm4apath)
	fmt.Println("param2:", programdir+"ffmpeg\\ffmpeg.exe", "-y", "-i", fpath, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-movflags", "faststart", outm4apath)
	cmd := exec.Command(programdir+"ffmpeg\\ffmpeg.exe", "-y", "-i", fpath, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-movflags", "faststart", outm4apath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	cpuusemap.Store(cpuid, "0")
}

var pathcnt sync.Map
var pathcntm sync.Mutex

func newcvtthread1(cpuid, startpos, inputpath, lenparam, lenstr, tracktitle, trackperformer, filename, cuetitle, cuedate, tracktag, outm4apath string, trackcnt int) {
	fmt.Println("convert to:", outm4apath)
	fmt.Println("parm1:", programdir+"ffmpeg_new\\ffmpeg.exe", "-y", "-ss", startpos, "-i", inputpath, lenparam, lenstr, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-metadata", "title="+tracktitle, "-metadata", "artist="+trackperformer, "-metadata", "comment="+filename, "-metadata", "album="+cuetitle, "-metadata", "date="+cuedate, "-metadata", "track="+tracktag, "-movflags", "faststart", outm4apath)
	_, err2 := os.Stat(outm4apath)
	if err2 != nil {
		cmd := exec.Command(programdir+"ffmpeg_new\\ffmpeg.exe", "-y", "-ss", startpos, "-i", inputpath, lenparam, lenstr, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-metadata", "title="+tracktitle, "-metadata", "artist="+trackperformer, "-metadata", "comment="+filename, "-metadata", "album="+cuetitle, "-metadata", "date="+cuedate, "-metadata", "track="+tracktag, "-movflags", "faststart", outm4apath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	fi, err := os.Stat(outm4apath)
	if err == nil && fi.Size() > 40 {
		pathcntm.Lock()
		value, bvalue := pathcnt.LoadOrStore(inputpath, 1)
		if bvalue {
			pathcnt.Store(inputpath, value.(int)+1)
		}
		pathcntm.Unlock()
		fmt.Println(value.(int), trackcnt)
		if value.(int)+1 == trackcnt || trackcnt == 1 {
			os.Remove(inputpath)
		}
	} else {
		value, bvalue := pathcnt.LoadOrStore(inputpath, 1)
		if bvalue {
			pathcnt.Store(inputpath, value.(int)-1)
		}
	}
	cpuusemap.Store(cpuid, "0")
}

func newcvtthread_lasttrack(cpuid, startpos, inputpath, tracktitle, trackperformer, filename, cuetitle, cuedate, tracktag, outm4apath string, trackcnt int) {
	fmt.Println("convert to:", outm4apath)
	fmt.Println("parm1:", programdir+"ffmpeg_new\\ffmpeg.exe", "-y", "-ss", startpos, "-i", inputpath, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-metadata", "title="+tracktitle, "-metadata", "artist="+trackperformer, "-metadata", "comment="+filename, "-metadata", "album="+cuetitle, "-metadata", "date="+cuedate, "-metadata", "track="+tracktag, "-movflags", "faststart", outm4apath)
	_, err2 := os.Stat(outm4apath)
	if err2 != nil {
		cmd := exec.Command(programdir+"ffmpeg_new\\ffmpeg.exe", "-y", "-ss", startpos, "-i", inputpath, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-metadata", "title="+tracktitle, "-metadata", "artist="+trackperformer, "-metadata", "comment="+filename, "-metadata", "album="+cuetitle, "-metadata", "date="+cuedate, "-metadata", "track="+tracktag, "-movflags", "faststart", outm4apath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	fi, err := os.Stat(outm4apath)
	if err == nil && fi.Size() > 40 {
		pathcntm.Lock()
		value, bvalue := pathcnt.LoadOrStore(inputpath, 1)
		if bvalue {
			pathcnt.Store(inputpath, value.(int)+1)
		}
		pathcntm.Unlock()
		if value.(int)+1 == trackcnt || trackcnt == 1 {
			os.Remove(inputpath)
		}
	} else {
		value, bvalue := pathcnt.LoadOrStore(inputpath, 1)
		if bvalue {
			pathcnt.Store(inputpath, value.(int)-1)
		}
	}
	cpuusemap.Store(cpuid, "0")
}

func newcvtthread2(cpuid, fpath, outm4apath string) {
	fmt.Println("convert to:", outm4apath)
	fmt.Println("parm3:", programdir+"ffmpeg_new\\ffmpeg.exe", "-y", "-i", fpath, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-movflags", "faststart", outm4apath)
	cmd := exec.Command(programdir+"ffmpeg_new\\ffmpeg.exe", "-y", "-i", fpath, "-map", "0:0", "-vn", "-b:a:0", rate, "-ac:a:0", "2", "-ar:a:0", "44100", "-movflags", "faststart", outm4apath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	fi, err := os.Stat(outm4apath)
	if err == nil && fi.Size() > 40 {
		os.Remove(fpath)
	}
	cpuusemap.Store(cpuid, "0")
}

func MoveDir(newdir, olddir string) {
	if len(olddir) > 0 && olddir[len(olddir)-1] != '\\' {
		olddir += "\\"
	}
	if len(newdir) > 0 && newdir[len(newdir)-1] != '\\' {
		newdir += "\\"
	}
	allname, err := ioutil.ReadDir(olddir)
	if err == nil {
		//do find one
		for i := 0; i < len(allname); i++ {
			if allname[i].IsDir() {
				os.Mkdir(newdir+allname[i].Name(), 0666)
				MoveDir(newdir+allname[i].Name(), olddir+allname[i].Name())
			} else {
				erro := os.Rename(olddir+allname[i].Name(), newdir+allname[i].Name())
				if erro != nil {
					panic(erro)
				}
			}
		}
	}
}

func DeleteDir(dirpath string) {
	if len(dirpath) > 0 && dirpath[len(dirpath)-1] != '\\' {
		dirpath += "\\"
	}
	allname, err := ioutil.ReadDir(dirpath)
	if err == nil {
		//do find one
		for i := 0; i < len(allname); i++ {
			if allname[i].IsDir() {
				DeleteDir(dirpath + allname[i].Name())
			} else {
				erro := os.Remove(dirpath + allname[i].Name())
				if erro != nil {
					panic(erro)
				}
			}
		}
	}
	os.Remove(dirpath)
}

func fn(fname string) string {
	fname = strings.Replace(fname, "\r", "", -1)
	fname = strings.Replace(fname, "\n", "", -1)
	fname = strings.Replace(fname, "\t", "_", -1)
	fname = strings.Replace(fname, "*", "_", -1)
	fname = strings.Replace(fname, "?", "_", -1)
	fname = strings.Replace(fname, "/", "_", -1)
	fname = strings.Replace(fname, "\\", "_", -1)
	fname = strings.Replace(fname, "|", "_", -1)
	fname = strings.Replace(fname, "\"", "_", -1)
	fname = strings.Replace(fname, "<", "_", -1)
	fname = strings.Replace(fname, ">", "_", -1)
	fname = strings.Replace(fname, ":", "_", -1)
	return fname
}

func DirFound(collectiondir string) bool {
	//fmt.Println("collectiondir:", collectiondir)
	if strings.Index(collectiondir, "FlacApeExtractTemp") != -1 {
		return false
	}
	if collectiondir[len(collectiondir)-1:] != "\\" {
		collectiondir += "\\"
	}
	allname, err := ioutil.ReadDir(collectiondir)
	if err == nil {
		//do find one
		for i := 0; i < len(allname); i++ {
			fname := allname[i].Name()
			if fname == "FlacApeExtractTemp" {
				continue
			}
			fpath := collectiondir + fname
			fmt.Println("fname:", fname)
			if allname[i].IsDir() {
				DirFound(fpath)
			} else if strings.LastIndex(fname, ".") != -1 {
				extstr := strings.ToLower(fname[strings.LastIndex(fname, "."):])
				switch extstr {
				case ".rar", ".zip", ".7z", ".iso":
					//uncompress
					collectiond := fpath
					collectiond = collectiond[0:strings.LastIndex(collectiond, ".")]
					fmt.Println("extract:", fpath)
					//param := []string{"C:\\Program Files\\7-Zip\\7z.exe", "x", "-y", "\"-o" + collectiond + "\"", curfilename}
					//fmt.Println(strings.Join(param, " "))
					collectiond2 := strings.Replace(collectiond, "(ED2000.COM)", "", -1)
					if len(collectiond2)-1 > 0 && collectiond2[len(collectiond2)-1] == '.' {
						collectiond2 = collectiond2[:len(collectiond2)-1]
					}
					_, szfie := os.Stat("C:\\Program Files\\7-Zip\\7z.exe")
					if szfie != nil {
						panic("7z not found at:" + "C:\\Program Files\\7-Zip\\7z.exe")
					}
					cmd := exec.Command("C:\\Program Files\\7-Zip\\7z.exe", "x", "-y", "-o"+collectiond2+"", fpath)
					//cmd.Stdout = os.Stdout
					//cmd.Stderr = os.Stderr
					//cmd.Run()
					//stdout, _ := cmd.StdoutPipe()
					stderr, _ := cmd.StderrPipe()
					cmd.Start()
					//stdoutbt, _ := ioutil.ReadAll(stdout)
					stderrbt, _ := ioutil.ReadAll(stderr)
					cmd.Wait()
					// ioutil.WriteFile("7z_error.log", stderrbt, 0666)
					// if len(stderrbt) == 0 {
					// 	panic("run 7z command empty return error")
					// }
					dirsize := toolfunc.DirGetSize(".*", collectiond2)
					fmt.Println("codifi size", dirsize)
					if dirsize == 0 {
						toolfunc.AppendFile("7zExtractErrorFiles.txt", []byte(fpath+"\n"), 0666)
						//ioutil.WriteFile("7z_error.log", []byte(fpath), 0666)
						//panic("7z extract error:" + fpath)
						toolfunc.RemoveDirAll(collectiond2)
						continue
					} else {
						if strings.Index(string(stderrbt), "ERROR") == -1 {
							os.Remove(fpath)
						}
					}
					colpath := collectiondir + "\\FlacApeExtractTemp"
					os.Mkdir(colpath, 0666)
					os.Rename(fpath, colpath+"\\"+allname[i].Name())
					//is extractdir has one dir move to top
					subdircnt := 0
					subfilecnt := 0
					subdirname := ""
					allname3, err3 := ioutil.ReadDir(collectiond2)
					if err3 == nil {
						//do find one
						for i3 := 0; i3 < len(allname3); i3++ {
							if allname3[i3].IsDir() {
								subdirname = allname3[i3].Name()
								subdircnt++
							} else {
								subfilecnt++
							}
						}
					}
					if subfilecnt == 0 && subdircnt == 1 {
						fmt.Println("rename dir:", collectiond2+"\\"+subdirname, collectiond2)
						MoveDir(collectiond2, collectiond2+"\\"+subdirname)
						os.Remove(collectiond2 + "\\" + subdirname)
					}

					//scan dir for standard
					toolfunc.DirTextFileConvert(".*[.]cue$", collectiond2, "UTF-8")
					// apppdir := toolfunc.AppDir()
					// cmd = exec.Command(apppdir+"ConvertTextFileToOtherCode.exe", ".*[.]cue$", collectiond2+"")
					// cmd.Stdout = os.Stdout
					// cmd.Stderr = os.Stderr
					// cmd.Run()
					// cmd.Wait()
					ScanDirForStd(collectiond2 + "")
					DirFound(collectiond2 + "")
				case ".flac", ".ape", ".wav", ".m4a":
					//exists cue?
					cuepath := fpath[0:strings.LastIndex(fpath, ".")] + ".cue"
					fi2, err2 := os.Stat(cuepath)
					if err2 == nil && !fi2.IsDir() {
						kkpath := fpath[:strings.LastIndex(fpath, "\\")+1]
						tempath := kkpath + "FlacApeExtractTemp"
						tempf, tempferr := os.Stat(tempath)

						kkpathfls, kkpatherr := ioutil.ReadDir(kkpath)
						kkpathfcnt := 0
						for _, kkpathfname := range kkpathfls {
							kkpathfname2 := strings.ToLower(kkpathfname.Name())
							if strings.HasSuffix(kkpathfname2, ".ape") || strings.HasSuffix(kkpathfname2, ".flac") || strings.HasSuffix(kkpathfname2, ".wav") || strings.HasSuffix(kkpathfname2, ".m4a") {
								kkpathfcnt += 1
							}
						}

						m4aExistsCnt := 0
						//cue exists;parse cue and convert
						cuedata, err := ioutil.ReadFile(cuepath)
						if err == nil && strings.Trim(string(cuedata), " \r\n\t") != "" {
							charsetname := charset.DetectCharset(cuedata)
							if charsetname != "UTF-8" {
								cuedata = []byte(charset.Convert(string(cuedata), charsetname, "UTF-8"))
							}
							cuedata2 := string(cuedata)
							cuedatals := strings.Split(cuedata2, "\n")
							binhead := true
							stdname := fpath[strings.LastIndex(fpath, "\\")+1:]
							stdname = stdname[0:strings.LastIndex(stdname, ".")]
							trackcnt := 0
							for ind, _ := range cuedatals {
								cuedatals[ind] = strings.Trim(cuedatals[ind], "\r\n\t ")
								if len(cuedatals[ind]) >= 5 && cuedatals[ind][0:5] == "TRACK" {
									trackcnt++
								}
							}
							var curtrack int = 1
							var cuetitle, cuedate string
							for ind, _ := range cuedatals {
								cuedatals[ind] = strings.Trim(cuedatals[ind], "\r\n\t ")
								if len(cuedatals[ind]) >= 5 && cuedatals[ind][0:5] != "TRACK" && binhead {
									if cuedatals[ind][0:4] == "FILE" && strings.LastIndex(cuedatals[ind], ".") != -1 {
										cuedatals[ind] = cuedatals[ind][0:strings.Index(cuedatals[ind], "\"")+1] + stdname + cuedatals[ind][strings.LastIndex(cuedatals[ind], "."):]
									}
									if len(cuedatals[ind]) > 5 && cuedatals[ind][0:5] == "TITLE" {
										cuetitle = strings.Trim(cuedatals[ind][5:], " \t\r\n")
										if len(cuetitle) > 0 && cuetitle[0] == '"' {
											cuetitle = cuetitle[1:]
										}
										if len(cuetitle) > 0 && cuetitle[len(cuetitle)-1] == '"' {
											cuetitle = cuetitle[0 : len(cuetitle)-1]
										}
									}
									if cuedatals[ind][0:4] == "DATE" {
										cuedate = strings.Trim(cuedatals[ind][4:], " \t\r\n")
										if len(cuedate) > 0 && cuedate[0] == '"' {
											cuedate = cuedate[1:]
										}
										if len(cuedate) > 0 && cuedate[len(cuedate)-1] == '"' {
											cuedate = cuedate[0 : len(cuedate)-1]
										}
									}
								} else {
									binhead = false
									if len(cuedatals[ind]) >= 5 && cuedatals[ind][0:5] == "TRACK" {
										//parse info and convert
										ind2 := ind
										var musicstartpos float32
										var tracktitle, performer string
										for ind2+1 < len(cuedatals) {
											ind2++
											if len(cuedatals[ind2]) >= 8 && cuedatals[ind2][0:8] == "INDEX 00" {

											}
											if len(cuedatals[ind2]) >= 8 && cuedatals[ind2][0:8] == "INDEX 01" {
												timestr := strings.Trim(cuedatals[ind2][9:], " \t\r\n")
												timestrls := strings.Split(timestr, ":")
												if len(timestrls) >= 3 {
													timeminute, _ := strconv.ParseFloat(timestrls[0], 32)
													timesec, _ := strconv.ParseFloat(timestrls[1], 32)
													timeframe, _ := strconv.ParseFloat(timestrls[2], 32)
													musicstartpos = float32(timeminute*60 + timesec + timeframe/75)
												}
											}
											if len(cuedatals[ind2]) > 5 && cuedatals[ind2][0:5] == "TITLE" {
												tracktitle = strings.Trim(cuedatals[ind2][5:], " \t\r\n")
												if len(tracktitle) > 0 && tracktitle[0] == '"' {
													tracktitle = tracktitle[1:]
												}
												if len(tracktitle) > 0 && tracktitle[len(tracktitle)-1] == '"' {
													tracktitle = tracktitle[0 : len(tracktitle)-1]
												}
											}
											if len(cuedatals[ind2]) > 9 && cuedatals[ind2][0:9] == "PERFORMER" {
												performer = strings.Trim(cuedatals[ind2][9:], " \t\r\n")
												if len(performer) > 0 && performer[0] == '"' {
													performer = performer[1:]
												}
												if len(performer) > 0 && performer[len(performer)-1] == '"' {
													performer = performer[0 : len(performer)-1]
												}
											}
											if ind2 >= len(cuedatals) || len(cuedatals[ind2]) >= 5 && cuedatals[ind2][0:5] == "TRACK" {
												break
											}
										}
										var startpos2, musicstartpos2 float32
										for ind2+1 < len(cuedatals) {
											ind2++
											if len(cuedatals[ind2]) > 8 && cuedatals[ind2][0:8] == "INDEX 00" {
												timestr := strings.Trim(cuedatals[ind2][8:], " \t\r\n")
												timestrls := strings.Split(timestr, ":")
												if len(timestrls) >= 3 {
													timeminute, _ := strconv.ParseFloat(timestrls[0], 32)
													timesec, _ := strconv.ParseFloat(timestrls[1], 32)
													timeframe, _ := strconv.ParseFloat(timestrls[2], 32)
													startpos2 = float32(timeminute*60 + timesec + timeframe/75)
												}
											}
											if len(cuedatals[ind2]) > 8 && cuedatals[ind2][0:8] == "INDEX 01" {
												timestr := strings.Trim(cuedatals[ind2][8:], " \t\r\n")
												timestrls := strings.Split(timestr, ":")
												if len(timestrls) >= 3 {
													timeminute, _ := strconv.ParseFloat(timestrls[0], 32)
													timesec, _ := strconv.ParseFloat(timestrls[1], 32)
													timeframe, _ := strconv.ParseFloat(timestrls[2], 32)
													musicstartpos2 = float32(timeminute*60 + timesec + timeframe/75)
												}
											}
											if ind2 >= len(cuedatals) || len(cuedatals[ind2]) >= 5 && cuedatals[ind2][0:5] == "TRACK" {
												break
											}
										}
										lenneedadd02 := false
										if musicstartpos-0.05 > 0 {
											musicstartpos -= 0.05
											lenneedadd02 = true
										}

										var musiclen float32
										if startpos2 > 0 && (musicstartpos2 == 0 || musicstartpos2 > 0 && startpos2 > musicstartpos2) {
											musiclen = startpos2 - musicstartpos
											//fmt.Println("musiclen = startpos2 - musicstartpos:", musiclen, startpos2, musicstartpos2,musicstartpos)
										} else if musicstartpos2 > 0 && (startpos2 == 0 || startpos2 > 0 && musicstartpos2 > startpos2) {
											musiclen = musicstartpos2 - musicstartpos
											//fmt.Println("musiclen = musicstartpos2 - musicstartpos:", musiclen, musicstartpos2, startpos2,musicstartpos)
										}
										if lenneedadd02 {
											musiclen += 0.1
										}
										musiclen += 0.05
										if musiclen <= 0.3 {
											musiclen = 0
										}
										//fmt.Println("lenparam:", lenparam, " lenstr:", lenstr,"trackcnt:",trackcnt,curtrack)

										ind = ind2
										//
										outm4apath := fpath[0:strings.LastIndex(fpath, "\\")] + "\\" + strconv.FormatInt(int64(curtrack), 10) + "." + fn(tracktitle) + ".m4a"
										_, m4afie := os.Stat(outm4apath)
										if m4afie == nil {
											m4aExistsCnt += 1
										}
									}
								}
							}
						}

						if m4aExistsCnt == 0 && (tempferr == nil && tempf.IsDir() || kkpatherr == nil && kkpathfcnt > 1) {
							editionname := cuepath[strings.LastIndex(cuepath, "\\")+1 : strings.LastIndex(cuepath, ".")]
							newdirpath := kkpath + editionname + "\\"
							os.Mkdir(newdirpath, 0666)
							os.Rename(cuepath, kkpath+editionname+"\\"+cuepath[strings.LastIndex(cuepath, "\\")+1:])
							fmt.Println("Move cue to directory:", cuepath, kkpath+editionname+"\\"+cuepath[strings.LastIndex(cuepath, "\\")+1:])
							cuepath = kkpath + editionname + "\\" + cuepath[strings.LastIndex(cuepath, "\\")+1:]
							os.Rename(fpath, kkpath+editionname+"\\"+fpath[strings.LastIndex(fpath, "\\")+1:])
							fmt.Println("Move media to directory:", fpath, kkpath+editionname+"\\"+fpath[strings.LastIndex(fpath, "\\")+1:])
							fpath = kkpath + editionname + "\\" + fpath[strings.LastIndex(fpath, "\\")+1:]

						}

						//cue exists;parse cue and convert
						cuedata, err = ioutil.ReadFile(cuepath)
						if err == nil && strings.Trim(string(cuedata), " \r\n\t") != "" {
							charsetname := charset.DetectCharset(cuedata)
							if charsetname != "UTF-8" {
								cuedata = []byte(charset.Convert(string(cuedata), charsetname, "UTF-8"))
							}
							cuedata2 := string(cuedata)
							cuedatals := strings.Split(cuedata2, "\n")
							binhead := true
							stdname := fpath[strings.LastIndex(fpath, "\\")+1:]
							stdname = stdname[0:strings.LastIndex(stdname, ".")]
							trackcnt := 0
							for ind, _ := range cuedatals {
								cuedatals[ind] = strings.Trim(cuedatals[ind], "\r\n\t ")
								if len(cuedatals[ind]) >= 5 && cuedatals[ind][0:5] == "TRACK" {
									trackcnt++
								}
							}
							var curtrack int = 1
							var cuetitle, cuedate string
							for ind, _ := range cuedatals {
								cuedatals[ind] = strings.Trim(cuedatals[ind], "\r\n\t ")
								if len(cuedatals[ind]) >= 5 && cuedatals[ind][0:5] != "TRACK" && binhead {
									if cuedatals[ind][0:4] == "FILE" && strings.LastIndex(cuedatals[ind], ".") != -1 {
										cuedatals[ind] = cuedatals[ind][0:strings.Index(cuedatals[ind], "\"")+1] + stdname + cuedatals[ind][strings.LastIndex(cuedatals[ind], "."):]
									}
									if len(cuedatals[ind]) > 5 && cuedatals[ind][0:5] == "TITLE" {
										cuetitle = strings.Trim(cuedatals[ind][5:], " \t\r\n")
										if len(cuetitle) > 0 && cuetitle[0] == '"' {
											cuetitle = cuetitle[1:]
										}
										if len(cuetitle) > 0 && cuetitle[len(cuetitle)-1] == '"' {
											cuetitle = cuetitle[0 : len(cuetitle)-1]
										}
									}
									if cuedatals[ind][0:4] == "DATE" {
										cuedate = strings.Trim(cuedatals[ind][4:], " \t\r\n")
										if len(cuedate) > 0 && cuedate[0] == '"' {
											cuedate = cuedate[1:]
										}
										if len(cuedate) > 0 && cuedate[len(cuedate)-1] == '"' {
											cuedate = cuedate[0 : len(cuedate)-1]
										}
									}
								} else {
									binhead = false
									if len(cuedatals[ind]) >= 5 && cuedatals[ind][0:5] == "TRACK" {
										//parse info and convert
										ind2 := ind
										var musicstartpos float32
										var tracktitle, performer string
										for ind2+1 < len(cuedatals) {
											ind2++
											if len(cuedatals[ind2]) >= 8 && cuedatals[ind2][0:8] == "INDEX 00" {

											}
											if len(cuedatals[ind2]) >= 8 && cuedatals[ind2][0:8] == "INDEX 01" {
												timestr := strings.Trim(cuedatals[ind2][9:], " \t\r\n")
												timestrls := strings.Split(timestr, ":")
												if len(timestrls) >= 3 {
													timeminute, _ := strconv.ParseFloat(timestrls[0], 32)
													timesec, _ := strconv.ParseFloat(timestrls[1], 32)
													timeframe, _ := strconv.ParseFloat(timestrls[2], 32)
													musicstartpos = float32(timeminute*60 + timesec + timeframe/75)
												}
											}
											if len(cuedatals[ind2]) > 5 && cuedatals[ind2][0:5] == "TITLE" {
												tracktitle = strings.Trim(cuedatals[ind2][5:], " \t\r\n")
												if len(tracktitle) > 0 && tracktitle[0] == '"' {
													tracktitle = tracktitle[1:]
												}
												if len(tracktitle) > 0 && tracktitle[len(tracktitle)-1] == '"' {
													tracktitle = tracktitle[0 : len(tracktitle)-1]
												}
											}
											if len(cuedatals[ind2]) > 9 && cuedatals[ind2][0:9] == "PERFORMER" {
												performer = strings.Trim(cuedatals[ind2][9:], " \t\r\n")
												if len(performer) > 0 && performer[0] == '"' {
													performer = performer[1:]
												}
												if len(performer) > 0 && performer[len(performer)-1] == '"' {
													performer = performer[0 : len(performer)-1]
												}
											}
											if ind2 >= len(cuedatals) || len(cuedatals[ind2]) >= 5 && cuedatals[ind2][0:5] == "TRACK" {
												break
											}
										}
										var startpos2, musicstartpos2 float32
										for ind2+1 < len(cuedatals) {
											ind2++
											if len(cuedatals[ind2]) > 8 && cuedatals[ind2][0:8] == "INDEX 00" {
												timestr := strings.Trim(cuedatals[ind2][8:], " \t\r\n")
												timestrls := strings.Split(timestr, ":")
												if len(timestrls) >= 3 {
													timeminute, _ := strconv.ParseFloat(timestrls[0], 32)
													timesec, _ := strconv.ParseFloat(timestrls[1], 32)
													timeframe, _ := strconv.ParseFloat(timestrls[2], 32)
													startpos2 = float32(timeminute*60 + timesec + timeframe/75)
												}
											}
											if len(cuedatals[ind2]) > 8 && cuedatals[ind2][0:8] == "INDEX 01" {
												timestr := strings.Trim(cuedatals[ind2][8:], " \t\r\n")
												timestrls := strings.Split(timestr, ":")
												if len(timestrls) >= 3 {
													timeminute, _ := strconv.ParseFloat(timestrls[0], 32)
													timesec, _ := strconv.ParseFloat(timestrls[1], 32)
													timeframe, _ := strconv.ParseFloat(timestrls[2], 32)
													musicstartpos2 = float32(timeminute*60 + timesec + timeframe/75)
												}
											}
											if ind2 >= len(cuedatals) || len(cuedatals[ind2]) >= 5 && cuedatals[ind2][0:5] == "TRACK" {
												break
											}
										}
										lenneedadd02 := false
										if musicstartpos-0.05 > 0 {
											musicstartpos -= 0.05
											lenneedadd02 = true
										}

										var musiclen float32
										if startpos2 > 0 && (musicstartpos2 == 0 || musicstartpos2 > 0 && startpos2 > musicstartpos2) {
											musiclen = startpos2 - musicstartpos
											//fmt.Println("musiclen = startpos2 - musicstartpos:", musiclen, startpos2, musicstartpos2,musicstartpos)
										} else if musicstartpos2 > 0 && (startpos2 == 0 || startpos2 > 0 && musicstartpos2 > startpos2) {
											musiclen = musicstartpos2 - musicstartpos
											//fmt.Println("musiclen = musicstartpos2 - musicstartpos:", musiclen, musicstartpos2, startpos2,musicstartpos)
										}
										if lenneedadd02 {
											musiclen += 0.1
										}
										musiclen += 0.05
										if musiclen <= 0.3 {
											musiclen = 0
										}
										var lenparam string
										var lenstr string
										if musiclen > 0 {
											lenparam = "-t"
											lenstr = strconv.FormatFloat(float64(musiclen), 'f', 3, 32)
										}
										//fmt.Println("lenparam:", lenparam, " lenstr:", lenstr,"trackcnt:",trackcnt,curtrack)

										ind = ind2
										//
										outm4apath := fpath[0:strings.LastIndex(fpath, "\\")] + "\\" + strconv.FormatInt(int64(curtrack), 10) + "." + fn(tracktitle) + ".m4a"
										fi4, err4 := os.Stat(outm4apath)
										bconvert := true
										if err4 == nil && !fi4.IsDir() {
											os.Remove(outm4apath)
										}
										if bconvert == true {
											for true {
												bfoundcpu := false
												for cpui := 0; cpui < runtime.NumCPU(); cpui++ {
													cpuid := strconv.FormatInt(int64(cpui), 10)
													cpuval, _ := cpuusemap.LoadOrStore(cpuid, "0")
													if cpuval == "0" {
														cpuusemap.Store(cpuid, "1")
														bfoundcpu = true
														filename := fpath[strings.LastIndex(fpath, "\\")+1:]
														filename = filename[0:strings.LastIndex(filename, ".")]
														//ofi, ofierr := os.Stat(outm4apath)
														//if ofierr == nil && ofi.Size() <= 40 {
														fmt.Println("outm4apath:", outm4apath)
														os.Remove(outm4apath)
														//}

														pathsegs := toolfunc.SplitAny(outm4apath, "/\\")
														for pasegi := len(pathsegs) - 1; pasegi >= 1; pasegi-- {
															padir := strings.Join(pathsegs[:pasegi], "/")
															pasegpath := padir + "/" + pathsegs[len(pathsegs)-1]
															_, pafie := os.Stat(pasegpath)
															if pafie == nil {
																os.Remove(pasegpath)
															}
														}

														if curtrack < trackcnt {
															go newcvtthread1(cpuid, strconv.FormatFloat(float64(musicstartpos), 'f', 3, 32), fpath, lenparam, lenstr, tracktitle, performer, filename, cuetitle, cuedate, strconv.FormatInt(int64(curtrack), 10)+"/"+strconv.FormatInt(int64(trackcnt), 10), outm4apath, trackcnt)
														} else {
															go newcvtthread_lasttrack(cpuid, strconv.FormatFloat(float64(musicstartpos), 'f', 3, 32), fpath, tracktitle, performer, filename, cuetitle, cuedate, strconv.FormatInt(int64(curtrack), 10)+"/"+strconv.FormatInt(int64(trackcnt), 10), outm4apath, trackcnt)
														}
														break
													}
												}
												if bfoundcpu == true {
													break
												} else {
													time.Sleep(1 * time.Second)
												}
											}
										}
										curtrack++
									}
								}
							}
						}
					} else if extstr != ".m4a" {
						//
						outm4apath := fpath[0:strings.LastIndex(fpath, ".")] + ".m4a"
						fi4, err4 := os.Stat(outm4apath)
						bconvert := true
						if err4 == nil && !fi4.IsDir() {
							os.Remove(outm4apath)
						}
						if bconvert == true {
							for true {
								bfoundcpu := false
								for cpui := 0; cpui < runtime.NumCPU(); cpui++ {
									cpuid := strconv.FormatInt(int64(cpui), 10)
									cpuval, _ := cpuusemap.LoadOrStore(cpuid, "0")
									if cpuval == "0" {
										cpuusemap.Store(cpuid, "1")
										bfoundcpu = true
										os.Remove(outm4apath)
										go newcvtthread2(cpuid, fpath, outm4apath)
										break
									}
								}
								if bfoundcpu == true {
									break
								} else {
									time.Sleep(1 * time.Second)
								}
							}
						}
					}
				}
			}
		}
	}
	return false
}

func DirFindErrorM4a(collectiondir string) int {
	errcnt := 0
	if strings.Index(collectiondir, "FlacApeExtractTemp") != -1 {
		return errcnt
	}
	if collectiondir[len(collectiondir)-1:] != "\\" {
		collectiondir += "\\"
	}
	allname, err := ioutil.ReadDir(collectiondir)
	if err == nil {
		//do find one
		for i := 0; i < len(allname); i++ {
			fname := allname[i].Name()
			if fname == "FlacApeExtractTemp" {
				continue
			}
			fpath := collectiondir + fname
			//fmt.Println("fname:", fname)
			if allname[i].IsDir() {
				errcnt += DirFindErrorM4a(fpath)
			} else if strings.LastIndex(fname, ".") != -1 {
				extstr := strings.ToLower(fname[strings.LastIndex(fname, "."):])
				switch extstr {
				case ".m4a":
					outm4apath := fpath[0:strings.LastIndex(fpath, ".")] + ".m4a"
					fi4, err4 := os.Stat(outm4apath)
					if err4 == nil && !fi4.IsDir() {
						if fi4.Size() <= 40 {
							fmt.Println("error m4a:", outm4apath)
							os.Remove(outm4apath)
							errcnt += 1
						}
					}
				}
			}
		}
	}
	return errcnt
}

func cleardirapeflac(collectiondir string) {
	if collectiondir[len(collectiondir)-1:] != "\\" {
		collectiondir += "\\"
	}
	if strings.Index(collectiondir, "FlacApeExtractTemp") != -1 {
		return
	}
	allname, err := ioutil.ReadDir(collectiondir)
	if err == nil {
		//do find one
		for i := 0; i < len(allname); i++ {
			fname := allname[i].Name()
			if allname[i].IsDir() {
				cleardirapeflac(collectiondir + fname)
			} else if strings.LastIndex(fname, ".") != -1 {
				switch strings.ToLower(fname[strings.LastIndex(fname, "."):]) {
				case ".ape", ".flac", ".wav", ".m4a":
					file, err := os.OpenFile(collectiondir+fname, os.O_RDWR, 0666)
					h := sha1.New()
					if err == nil {
						buf := make([]byte, 1024*1024)
						for {
							length, err := file.Read(buf)
							if err != nil {
								if err == io.EOF {
									break
								} else {
								}
							}
							h.Write(buf[0:length])
						}
						file.Close()
					}
					bs := h.Sum(nil)
					sumhex := fmt.Sprintf("%x", bs)
					ioutil.WriteFile(collectiondir+fname+".sha1", []byte(sumhex), 0666)
					os.Remove(collectiondir + fname)
				}
			}
		}
	}
}

func r2k(str string) string {
	return strings.Replace(str, "(ED2000.COM)", "", -1)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("no collection directories!")
		return
	}

	programdir = os.Args[0]
	programdir = programdir[0 : strings.LastIndex(programdir, "\\")+1]

	confdata2, _ := ioutil.ReadFile(programdir + "FlacApeToM4aConverter.conf")
	confdata := string(confdata2)
	confls := strings.Split(confdata, "\n")
	for i := 0; i < len(confls); i++ {
		if len(confls[i]) > 5 && strings.Compare("rate=", confls[i][0:5]) == 0 {
			rate = strings.Trim(confls[i][5:], " \r\n\t")
		}
	}

	bvecompress := false
	for _, collectiondir := range os.Args {
		allname, err := ioutil.ReadDir(collectiondir)
		if err == nil {
			//do find one
			for i := 0; i < len(allname); i++ {
				curfilename := collectiondir + "\\" + allname[i].Name()
				if strings.LastIndex(curfilename, ".") == -1 {
					continue
				}
				extstr := strings.ToLower(curfilename[strings.LastIndex(curfilename, "."):])
				switch extstr {
				case ".zip", ".rar", ".7z":
					bvecompress = true
				}

			}
		}
	}
	if bvecompress {
		for _, collectiondir := range os.Args {
			allname, err := ioutil.ReadDir(collectiondir)
			if err == nil {
				//do find one
				for i := 0; i < len(allname); i++ {
					curfilename := collectiondir + "\\" + allname[i].Name()
					if strings.LastIndex(curfilename, ".") == -1 {
						continue
					}
					extstr := strings.ToLower(curfilename[strings.LastIndex(curfilename, "."):])
					switch extstr {
					case ".ape", ".flac", ".wav", ".m4a":
						//find .ape .ape.cue and rename
						cuefilepath := curfilename[0:strings.LastIndex(curfilename, ".")] + ".cue"
						fi2, err2 := os.Stat(cuefilepath)
						if err2 == nil && !fi2.IsDir() {
							apeflacdir := curfilename[0:strings.LastIndex(curfilename, ".")]
							os.Mkdir(r2k(apeflacdir), 0666)
							os.Rename(cuefilepath, r2k(apeflacdir+"\\"+cuefilepath[strings.LastIndex(cuefilepath, "\\")+1:]))
							os.Rename(curfilename, r2k(apeflacdir+"\\"+curfilename[strings.LastIndex(curfilename, "\\")+1:]))
						}
					}
				}
			}
		}
	}

	for _, collectiondir := range os.Args {
		if strings.Index(collectiondir, "FlacApeExtractTemp") != -1 {
			continue
		}
		fi, err := os.Stat(collectiondir)
		if err == nil && fi.IsDir() {
			if fi.Name() == "FlacApeExtractTemp" {
				continue
			}
			toolfunc.DirTextFileConvert(".*[.]cue$", collectiondir, "UTF-8")
			// cmd := exec.Command(programdir+"ConvertTextFileToOtherCode.exe", ".*[.]cue$", collectiondir)
			// cmd.Stdout = os.Stdout
			// cmd.Stderr = os.Stderr
			// cmd.Run()
			// cmd.Wait()
			ScanDirForStd(collectiondir)
			DirFound(collectiondir)
		}
	}

	//all done del ape flac
	for true {
		zerocpucnt := 0
		for cpui := 0; cpui < runtime.NumCPU(); cpui++ {
			cpuid := strconv.FormatInt(int64(cpui), 10)
			cpuval, _ := cpuusemap.LoadOrStore(cpuid, "0")
			if cpuval == "0" {
				zerocpucnt++
			}
		}
		if zerocpucnt == runtime.NumCPU() {
			break
		}
	}

	// allsubdir := toolfunc.GetDirSubdirname(os.Args[1])
	// for _, subdir := range allsubdir {
	// 	if strings.Index(subdir, "FlacApeExtractTemp") != -1 {
	// 		continue
	// 	}
	// 	toolfunc.DirTextFileConvert(".*[.]cue$", subdir, "UTF-8")
	// 	DirFound(subdir)
	// }

	errcnt := 0
	for _, collectiondir := range os.Args {
		_, err := ioutil.ReadDir(collectiondir)
		if err == nil {
			errcnt += DirFindErrorM4a(collectiondir)
		}
	}

	fmt.Println("errcnt:", errcnt)

	/*
		if errcnt == 0 {
			for argi, collectiondir := range os.Args {
				if argi == 0 {
					continue
				}
				fmt.Println("param co:", collectiondir)
				allname, err := ioutil.ReadDir(collectiondir)
				if err == nil {
					//do find one
					for i := 0; i < len(allname); i++ {
						if allname[i].IsDir() {
							if allname[i].Name() == "FlacApeExtractTemp" {
								continue
							}
							fmt.Println("cleardirapeflac", i, len(allname), collectiondir+"/"+allname[i].Name())
							cleardirapeflac(collectiondir + "/" + allname[i].Name())
						}

					}
				}
			}
		}
	*/

}
