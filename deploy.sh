raida0=("root" "fNCEAgKQaCZ2G43f" "87.120.8.249" "22")
raida1=("root" "fNC*-_{$DTU!;" "23.106.122.6" "22")
raida2=("root" "K2QSfUN9qdDr" "172.105.176.86" "22")
raida3=("root" "Ix9qwwsraWtW%#" "85.195.82.169" "22")
raida4=("root" "vjVw9vm8vhcv" "198.244.135.236" "22")
raida5=("root" "3WWsN(R!UHbR" "88.119.174.101" "22")
raida6=("root" "OqGXUOb9tzSk" "209.141.52.193" "22")
raida7=("root" "gyT6s6SECFS9$~" "179.43.175.35" "22")
raida8=("root" "BO33qW7$SQSW" "104.161.32.116" "22")
raida9=("root" "Lnc2gTyOsIt4" "66.172.11.25" "22")
raida10=("root" "EsrF87jLXcag" "194.29.186.69" "22")
raida11=("root" "E3fAgKQfNC#!$" "168.235.69.182" "22")
raida12=("root" "T%$QfNC*-_{" "5.230.67.199" "22")
raida13=("root" "N9qdDrK2N9qd" "167.88.15.117" "22")
raida14=("root" "$XVRO8C4zv457" "23.29.115.137" "22")
raida15=("root" "(!bsB6Ty5h!q" "66.29.143.85" "22")
raida16=("root" "!8HEOfQQnaUe" "185.99.133.110" "22")
raida17=("root" "G!q9aIc84hI3" "104.168.162.230" "22")
raida18=("ubuntu" "!dHIL7Spkw4I" "170.75.170.4" "22")
raida19=("root" "6s6CTyMHpGRB" "185.215.227.31" "22")
raida20=("root" "2ktSRFxu?DG9" "51.222.229.205" "22")
raida21=("root" "*c$Yfz7$b35k" "31.192.107.132" "22")
raida22=("root" "Spk$Yfz7w4I" "180.235.135.143" "22")
raida23=("root" "YfzZi&VTS*c$Y" "80.233.134.148" "22")
raida24=("root" "C9vW8bbT6sst" "147.182.249.132" "22")

ALL_RAIDA=(
  raida0[@]
  raida1[@]
  raida2[@]
  raida3[@]
  raida4[@]
  raida5[@]
  raida6[@]
  raida7[@]
  raida8[@]
  raida9[@]
  raida10[@]
  raida11[@]
  raida12[@]
  raida13[@]
  raida14[@]
  raida15[@]
  raida16[@]
  raida17[@]
  raida18[@]
  raida19[@]
  raida20[@]
  raida21[@]
  raida22[@]
  raida23[@]
  raida24[@]
)

COUNT=${#ALL_RAIDA[@]}
echo "${COUNT}"
for ((i=0; i<$COUNT; i++))
do 
	echo "*Deployment process started for raida $i"
	#sshpass -p ${!ALL_RAIDA[i]:1:1} ssh ${!ALL_RAIDA[i]:0:1}@${!ALL_RAIDA[i]:2:1} -o StrictHostKeyChecking=no -p ${!ALL_RAIDA[i]:3:1} "free -m"
	sshpass -p ${!ALL_RAIDA[i]:1:1} ssh ${!ALL_RAIDA[i]:0:1}@${!ALL_RAIDA[i]:2:1} -o StrictHostKeyChecking=no -p ${!ALL_RAIDA[i]:3:1} "echo 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCuDeSn+ibflUWQuc8dsw9FRyrzc07F08Rv0cHsAjsR3/p+f/5lGZYBA6dbR6BkljlqTfdFIcbzzVWTzydkcTEC1/QkFdgiSL4RiL9Z/FJXIDLnrQnr1oi6rsfnvX1nJ/8wANOLLwjBsRjeMIvFSXU13Bvfgx6/oFCGGHyLmlNtsVg+gcVmKiax+Y28fy2VeMAxvDSbRIP2EQ17XWVIxOhNMxBph938ErwNzi61ng+vqj59sdNKeNk4NOXh+tT1F4wb5O/uJvaDWw2OXgzIiR+2GQ4HsFOJASw3UKhiKft48oSC9cYjpZF3XRc+4XgnslRatt0lfTKsK5k3B9CTOu/v6A5Wix2vuLJ21jLJJ71xcUSxUHL2NfPOjBtuqwxsK2G7vK1OoR8nUI9i8mbpzXiZZOljxWXt2lanX5I2cnVFqym4pmHA8tY5aeH9VeSbNCxp1b7X5CLIyosN9PhrYvFOHEvA+xb/jB+msDdxF+UyMWYDCQJ4ACv99/3Lm6C8dAc= root@sean-Gazelle' >> /root/.ssh/authorized_keys"
	#sshpass -p ${!ALL_RAIDA[i]:1:1} ssh ${!ALL_RAIDA[i]:0:1}@${!ALL_RAIDA[i]:2:1} -o StrictHostKeyChecking=no -p ${!ALL_RAIDA[i]:3:1} "echo 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDl9FJLH02hNxaW/BfFqkEgId1eXTtJvTByu1sgqaLgLYUHEiurRDqPxrjdMzqAW+uaNJKuTfw7zHpf9Cq81buxiM5ABM3KJHl2Lh5JIVxpWskzxD9h/1j7HSLVAD2Ll4y2/Gg3PbOLhE63XmjGHiLURqwjQkxm7/roXC8N8G4xikOcDmY3JhKCJII0cvmZuhYdzhGrZUkY4JwTHb3qLbcvE60PW5T+YIsKQtJbFTvLk/jxvXHJcQnlNn3PLErqNGSRdRSbWyNxw8+P6aTLPbDJVKaP30J7IzbLAz7GleIQk7e2j/6VXtWf3aOg+Cz0kxFHF3qQwnqSeUH3j2qe7oyd' >> /root/.ssh/authorized_keys"
	#sshpass -p ${!ALL_RAIDA[i]:1:1} ssh -v ${!ALL_RAIDA[i]:0:1}@${!ALL_RAIDA[i]:2:1} -p ${!ALL_RAIDA[i]:3:1} "echo ${!ALL_RAIDA[i]:1:1} | sudo -iS pkill raida_server"
	#echo $'\n*copying files to server please wait ...*'
	#sshpass -p ${!ALL_RAIDA[i]:1:1} scp -P ${!ALL_RAIDA[i]:3:1} -pr /home/sean/raida_deployment/raida/ ${!ALL_RAIDA[i]:0:1}@${!ALL_RAIDA[i]:2:1}:/home/raida_deployment
	#echo $'\n*executing remote services...*'
	#sshpass -p ${!ALL_RAIDA[i]:1:1} ssh -v ${!ALL_RAIDA[i]:0:1}@${!ALL_RAIDA[i]:2:1} -p ${!ALL_RAIDA[i]:3:1} "echo ${!ALL_RAIDA[i]:1:1} | sudo -iS bash /home/raida_deployment/raida/remote_script.sh" >/dev/null 2>&1 & 
	echo $'\n*Deployment process completed*'
done




#Note:
#actually first time the scp needs the finger print to be added to the machine for the server with which its communicating
#so i removed the ssh pass for scp
#then for all the servers's it prompted me to add the finger print
#i said yes to all
#after that again i added the sspass keyword behind scp
#so that i will not ask password all the time
#Use below format
#do 
	#echo "*Deployment process started for raida $i"
	#sshpass -p ${!ALL_RAIDA[i]:1:1} ssh -v ${!ALL_RAIDA[i]:0:1}@${!ALL_RAIDA[i]:2:1} -p ${!ALL_RAIDA[i]:3:1} "echo ${!ALL_RAIDA[i]:1:1} | sudo -iS pkill raida_server"
	#echo $'\n*copying files to server please wait ...*'
	#scp -P ${!ALL_RAIDA[i]:3:1} -pr /home/sean/raida_deployment/raida/ ${!ALL_RAIDA[i]:0:1}@${!ALL_RAIDA[i]:2:1}:/home/raida_deployment
	#echo $'\n*executing remote services...*'
	#sshpass -p ${!ALL_RAIDA[i]:1:1} ssh -v ${!ALL_RAIDA[i]:0:1}@${!ALL_RAIDA[i]:2:1} -p ${!ALL_RAIDA[i]:3:1} "echo ${!ALL_RAIDA[i]:1:1} | sudo -iS bash #/home/raida_deployment/raida/remote_script.sh" >/dev/null 2>&1 & 
	#echo $'\n*Deployment process completed*'
#done

