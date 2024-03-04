echo "Creating qr folder if it doesn't exist"
mkdir qr
echo "[run.sh] Clearing qr folder"
rm qr/*
echo "[run.sh] Starting the main program"
go run server.go