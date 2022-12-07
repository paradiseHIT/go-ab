make clean && make
s=`date +%s`
./go-ab \
    -alsologtostderr \
    -request_file data.txt \
    -url http://127.0.0.1:8080/users  \
    -method POST \
    -request_num 5 \
    -thread_num 2 \
    -content_type "application/json" \
    -header "Authorization: xxx" \
    -header "A: B" \
    -v 3
e=`date +%s`
cost=`echo "$e - $s" | bc -l`
echo "cost $cost seconds"
