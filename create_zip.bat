@echo off
del MyBot.zip
"C:\Program Files\7-Zip\7z" a MyBot.zip MyBot.go
set dir=%CD%
cd %GOPATH%
"C:\Program Files\7-Zip\7z" a %dir%\MyBot.zip src\github.com\BenJuan26\hlt -x!src\github.com\BenJuan26\hlt\.git
cd %dir%