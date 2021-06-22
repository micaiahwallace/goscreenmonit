import { useEffect, useState, useCallback, useMemo } from 'react';
import request from 'axios';
import './App.css';
import { useRef } from 'react/cjs/react.production.min';

function App() {

  const [mons, setMons] = useState([])
  const [selected, setSelected] = useState(null);
  const [update, setUpdate] = useState(1);

  // Load monitors on initial load
  useEffect(() => {
    const interval = setInterval(() => {
      request("/monitors")
      .then(resp => {
        setMons(resp.data)
      })
    }, 2000)
    return () => clearInterval(interval)
  }, []);

  // Update selected monitor
  const setMon = useCallback((address) => {
    const mon = mons.find(m => m.address = address)
    setSelected(mon)
  }, [mons]);

  // update active screenshot
  useEffect(() => {
    if (false && selected) {
      // Toggle update increment between 1 and 2 instead of
      // incrementing indefinitely to avoid it running to Infinity
      const interval = setInterval(() => setUpdate(u => u == 1 ? 2 : 1), 950);
      return () => clearInterval(interval)
    }
  }, [selected]);

  // Get user to display in active monitor section
  const userTitle = useMemo(() => {
    return selected ? selected.user : null;
  }, [selected]);

  // Create image source urls
  const urls = [];
  if (selected) {
    for (let i = 0; i < selected.screenCount; i++) {
      urls.push(`/monitors/${selected.address}/${i}?r=${Math.random()}`);
    }
  }

  // canvas ref
  const canvas = useRef()

  // Connect to websocket to receive data
  useEffect(() => {
    if (selected) {
      const socket = new WebSocket(`wss://${window.location.host}/ws/${selected.address}/0`)
      socket.onmessage = (message) => {
        const ctx = canvas.current.getContext("2d")
        var img = new Image();
        img.onload = function() {
          canvas.current.width = img.width
          canvas.current.height = img.height
          ctx.drawImage(img, 0, 0)
        }
        img.src = URL.createObjectURL(message.data);
      }
      socket.onopen = () => {
        console.log("WS connected.");
      }
      socket.onclose = () => {
        console.log("WS disconnected.");
      }
      return () => socket.close();
    }
  }, [selected])

  // Generate image width
  const imWidth = selected ? Math.min(Math.max(100 / selected.screenCount, 50), 35) : 50;

  return (
    <div>
      <h1>Go Screen Monit</h1>
      <ul>
        {mons.map(mon => (
          <li key={mon.address}><a href="#" onClick={setMon.bind(null, mon.address)}>{mon.user} ({mon.host} - {mon.address})</a></li>
        ))}
      </ul>
      {
        userTitle && (
          <>
            <h2>Viewing Session: ({userTitle}) {selected && <button onClick={setSelected.bind(null, null)}>(Stop Viewing)</button>}</h2>
            {/*urls.map(url => (
              <img width={`${imWidth}%`} src={url} style={{ float: "left" }} />
            ))*/}
            <canvas ref={canvas} width="800" height="600"></canvas>
          </>
        )
      }
    </div>
  );
}

export default App;
