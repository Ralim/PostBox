FROM golang:1.19 AS build

WORKDIR /src/
# Pull in deps first to ease cache
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

# Use linkerflags to strip DWARF info
RUN CGO_ENABLED=0 go build -ldflags="-s -w"  -o /bin/PostBox

# Now create the runtime container from python base for nsz

FROM scratch

WORKDIR /PostBox/
COPY --from=build /bin/PostBox ./PostBox

ENTRYPOINT ["/PostBox/PostBox"]
