FROM golang:1.24.4-bookworm AS build
WORKDIR /src
COPY . .
RUN go build -o /bin/user ./main


FROM scratch
COPY --from=build /bin/user /bin/user
CMD [ "/bin/user" ]