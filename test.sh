make clean && make
s=`date +%s`
./go-ab \
    -alsologtostderr \
    -request_file data.txt \
    -url http://127.0.0.1:8080/users  \
    -method POST \
    -request_num 50000 \
    -thread_num 2 \
    -content_type "application/json" \
    -header "Authorization: xxx" \
    -header "A: B" \
    -duration 10s
e=`date +%s`
cost=`echo "$e - $s" | bc -l`
echo "cost $cost seconds"
