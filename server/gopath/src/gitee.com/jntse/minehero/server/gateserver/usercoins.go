package main

import (
	"gitee.com/jntse/gotoolkit/log"
)

func AutoGame(){
	for i := 1; i <= 11; i++{
		start := i
		path := make([]int32, 0)
		GoNext(1, int32(start), path)
	}
}

func IsLeftBorder (num int32) bool {
	if num == 1{
		return true
	}
	return false
}

func IsRightBorder (num int32) bool {
	if num == 11{
		return true
	}
	return false
}

func GoNext (row int32, num int32, path []int32){
	path = append(path, num)
	if row == 5{
		log.Info("球的最后路径:%v", path)
		return
	}
	var newnum int32
	if row % 2 != 0{
		if IsLeftBorder(num) {
			newnum = num
			GoNext(row+1, newnum, path)
		}else if IsRightBorder(num) {
			newnum = num-1
			GoNext(row+1, newnum, path)
		}else {
			newnum = num
			var newpath = make([]int32, len(path))
			copy(newpath, path)
			GoNext(row+1, newnum, newpath)
			newnum = num-1
			var newpath1 = make([]int32, len(path))
			//log.Info("球的中途路径1:%v", path)
			copy(newpath1, path)
			//log.Info("球的中途路径2:%v copied=%d", newpath1, cc)
			GoNext(row+1, newnum, newpath1)
		}
	}else{
		newnum = num
		var newpath = make([]int32, len(path))
		copy(newpath, path)
		GoNext(row+1, newnum, newpath)
		newnum = num+1
		var newpath1 = make([]int32, len(path))
		//log.Info("球的中途路径1:%v", path)
		copy(newpath1, path)
		//log.Info("球的中途路径2:%v copied=%d", newpath1, cc)
		GoNext(row+1, newnum, newpath1)
	}
}
