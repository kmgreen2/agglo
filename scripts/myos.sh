if [[ -n `echo $OSTYPE | grep "linux"` ]]; then
    echo "linux"
elif [[ -n `echo $OSTYPE | grep "darwin"` ]]; then
    echo "darwin"
else
    echo "notavailableyet"
fi
