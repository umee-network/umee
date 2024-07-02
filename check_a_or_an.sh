#!/bin/bash

echo 123


# 不能是an
not_an_array=(user utility unique useful url union uvarint uniform usage uid unified utilization unit unix utilized usual uintptr uint)


for item in ${not_an_array[@]}
do
  ag "an "$item
done


echo 
echo "========================"
echo 





# 不能是a
not_a_array=(http heir herb honest honesty honorable hour honor MBA FBI herb hypothesis historic hourglass xml)


for item in ${not_a_array[@]}
do
  ag "a "$item
done
