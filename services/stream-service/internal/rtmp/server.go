package rtmp

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/config"
	"github.com/nareix/joy5/format/flv"
	"github.com/nareix/joy5/format/rtmp"
)

type Handler interface {
	OnPublish(streamKey string) (io.WriteCloser, error)
	OnClose(streamKey string)
}

type Server struct {
	cfg     *config.RTMPConfig
	handler Handler
}

func NewServer(cfg *config.RTMPConfig, handler Handler) *Server {
	return &Server{cfg: cfg, handler: handler}
}

func (s *Server) Listen() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s.cfg.Port))
	if err != nil {
		return fmt.Errorf("rtmp listen :%s: %w", s.cfg.Port, err)
	}
	log.Printf("RTMP server listening on :%s", s.cfg.Port)

	srv := rtmp.NewServer()

	srv.HandleConn = func(c *rtmp.Conn, nc net.Conn) {
		defer nc.Close()

		if !c.Publishing {
			return
		}

		// Extract stream key from URL path
		// OBS sends: rtmp://host:1935/live/<stream_key>
		streamKey := c.URL.Path
		if len(streamKey) > 0 && streamKey[0] == '/' {
			streamKey = streamKey[1:]
		}
		if len(streamKey) > 5 && streamKey[:5] == "live/" {
			streamKey = streamKey[5:]
		}

		log.Printf("[RTMP] publish start: key=%s remote=%s", streamKey, nc.RemoteAddr())

		// Get the ffmpeg stdin pipe from the service layer
		pipe, err := s.handler.OnPublish(streamKey)
		if err != nil {
			log.Printf("[RTMP] rejected key=%s: %v", streamKey, err)
			return
		}
		defer pipe.Close()

		// Write FLV header to the pipe so ffmpeg knows the format
		muxer := flv.NewMuxer(pipe)

		// Read RTMP packets and forward them as FLV to ffmpeg
		for {
			pkt, err := c.ReadPacket()
			if err != nil {
				break
			}
			if err := muxer.WritePacket(pkt); err != nil {
				log.Printf("[RTMP] write to ffmpeg pipe failed: %v", err)
				break
			}
		}

		log.Printf("[RTMP] publish end: key=%s", streamKey)
		s.handler.OnClose(streamKey)
	}

	for {
		nc, err := listener.Accept()
		if err != nil {
			return err
		}
		go srv.HandleNetConn(nc)
	}
}
