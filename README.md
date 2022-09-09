
github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect \
github.com/apache/thrift v0.0.0-20161221203622-b2a4d4ae21c7 // indirect


go build -o los-workflow *.go


# first time build
git submodule update --init\

# create database schema
SQL_USER=root SQL_PASSWORD=password make install-schema-mysql\


# start cadence
cd cadence\
./cadence-server --zone mysql start \

# register samples-domain
./cadence --do samples-domain d re \

# start cadence-web
cd cadence-web\
npm run dev\
http://localhost:8088/domains/samples-domain




# build los workflow project
go build -o los-workflow *.go

# start los workflow
export MONGO_DATABASE=test \
export MONGO_URI=mongodb://localhost:27017/test 

go run *.go

# start mock server
export MONGO_DATABASE=test \
export MONGO_URI=mongodb://localhost:27017/test \
export RABBITMQ_URI="amqp://user:password@localhost:5672/" \
export RABBITMQ_QUEUE=nlos

go run ./los-api-server  *.go


# generate go grpc with buf
cd  los-api-server \
buf generate  losapis


