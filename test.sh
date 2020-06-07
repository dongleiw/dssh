# 测试
# 参数
#	1: 一台空闲机器ip. 用于测试

tmpdir=./test.tmpdir
mkdir -p $tmpdir

export test_another_host=$1

srcfiles=$(ls *.go|grep -v -E '_test.go$')
idx=1
for testfile in $(ls *_test.go); do
	echo "#######test $testfile"
	go test -c -count 1 $testfile $srcfiles	 -o $tmpdir/a.$idx
	if [ $? -ne 0 ];then
		exit 1
	fi
	$tmpdir/a.$idx
	((idx++))
done
