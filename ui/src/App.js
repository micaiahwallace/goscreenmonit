import { useEffect, useState, useCallback, useMemo } from 'react';
import request from 'axios';
import './App.css';

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
    if (selected) {
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
            {urls.map(url => (
              <img width={`${imWidth}%`} src={url} style={{ float: "left" }} />
            ))}
          </>
        )
      }
    </div>
  );
}

export default App;
