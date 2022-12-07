make clean && make
s=`date +%s`
./go-ab -log_dir=log -alsologtostderr -request_file data.txt -url http://127.0.0.1:8080/users  -method POST -request_num 500 -thread_num 2
e=`date +%s`
cost=`echo "$e - $s" | bc -l`
echo "cost $cost seconds"
