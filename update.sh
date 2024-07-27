EXE=gocuuull
go build -o $EXE
cp $EXE ~/bin
scp $EXE data:bin
