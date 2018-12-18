#!/bin/sh
oldstr=$1
newstr=$2

#如果参数3空，取当前目录
if [ -z $3]
then
    workpath=./
else
    workpath=$3
fi

sed -i "s/$oldstr/$newstr/g" `grep "$oldstr" -rl "$workpath"`
echo "替换字符串:" $oldstr "-->" $newstr
echo "替换目录:" $workpath
echo "替换完成"
