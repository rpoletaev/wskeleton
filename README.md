# wskeleton
Just allow write app logic over websocket realisation

```Go
import "github.com/rpoletaev/wskeleton"

// Application logic for Hub
type RoomHeader struct {
  ID bson.ObjectID
  Name string
}


type ChatRoom {
  RoomHeader
  wskeleton.Hub
}

func main(){
  cf := &wskeleton.HubConfig{MaxArchiveSize: 20}
  room := &ChatRoom{
    RoomHeader: GetRoomHeaderFromDB(),
    Hub: wskeleton.CreateHub(cf)
  }
  
  room.Run()
  
}
```
