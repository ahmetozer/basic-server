FROM golang:1.16
WORKDIR $GOPATH/src/github.com/ahmetozer/basic-server/
COPY go.mod ./
RUN go mod download

COPY config ./config
COPY pkg ./pkg
COPY .git ./.git
COPY *.go .
RUN export GIT_COMMIT=$(git rev-list -1 HEAD) && \
    export GIT_TAG=$(git tag | tail -1) && \
    export GIT_URL=$(git config --get remote.origin.url) && \
    CGO_ENABLED=0 go build -v -ldflags="-X 'main.GitUrl=$GIT_URL' -X 'main.GitTag=$GIT_TAG' -X 'main.GitCommit=$GIT_COMMIT' -X 'main.BuildTime=$(date -Isecond)' -X 'main.RunningEnv=container'" -o /app/basic-server

RUN export DEBIAN_FRONTEND=noninteractive && apt update && apt install -y libcap2-bin
RUN setcap CAP_NET_BIND_SERVICE=+eip /app/basic-server

RUN echo 'root:x:0:0:root:/root:/bin/ash\n\
bin:x:1:1:bin:/bin:/\n\
daemon:x:2:2:daemon:/sbin:/\n\
adm:x:3:4:adm:/var/adm:/ \n\
lp:x:4:7:lp:/var/spool/lpd:/\n\
sync:x:5:0:sync:/sbin:/bin/sync\n\
shutdown:x:6:0:shutdown:/sbin:/sbin/shutdown\n\
halt:x:7:0:halt:/sbin:/sbin/halt\n\
mail:x:8:12:mail:/var/mail:/\n\
news:x:9:13:news:/usr/lib/news:/\n\
uucp:x:10:14:uucp:/var/spool/uucppublic:/\n\
operator:x:11:0:operator:/root:/\n\
man:x:13:15:man:/usr/man:/\n\
postmaster:x:14:12:postmaster:/var/mail:/\n\
cron:x:16:16:cron:/var/spool/cron:/\n\
ftp:x:21:21::/var/lib/ftp:/\n\
sshd:x:22:22:sshd:/dev/null:/\n\
at:x:25:25:at:/var/spool/cron/atjobs:/\n\
squid:x:31:31:Squid:/var/cache/squid:/\n\
xfs:x:33:33:X Font Server:/etc/X11/fs:/\n\
games:x:35:35:games:/usr/games:/\n\
cyrus:x:85:12::/usr/cyrus:/\n\
vpopmail:x:89:89::/var/vpopmail:/\n\
ntp:x:123:123:NTP:/var/empty:/\n\
smmsp:x:209:209:smmsp:/var/spool/mqueue:/\n\
guest:x:405:100:guest:/dev/null:/\n\
nobody:x:65534:65534:Nobody:/:' > /app/passwd.minimal && \
mkdir -p /app/tmp/cert &&  chmod -R ugo+rw /app/tmp


FROM scratch
USER nobody
COPY config /config
COPY --from=0  /app/basic-server /basic-server
COPY --from=0  /app/passwd.minimal /etc/passwd
COPY --from=0  /app/tmp /tmp
LABEL org.opencontainers.image.source="https://github.com/ahmetozer/basic-server"
ENTRYPOINT [ "/basic-server" ]