import { useEffect, useState, useCallback } from 'react';
import request from 'axios';
import './App.css';

function App() {

  const [mons, setMons] = useState([])
  const [selected, setSelected] = useState(null);
  const [imgSrc, setSrc] = useState("");

  // Load monitors on initial load
  useEffect(() => {
    request("/monitors")
    .then(resp => {
      setMons(resp.data)
    })
  }, []);

  // Update selected monitor
  const setMon = useCallback((address) => {
    setSelected(address)
  }, []);

  // update active screenshot
  useEffect(() => {
    if (selected !== null) {
      const timer = setInterval(() => {
        setSrc(`/monitors/${selected}?${Math.random()}`)
      }, 50);
      return () => clearInterval(timer)
    }
  }, [selected]);

  return (
    <div>
      <h1>S1 Monitor</h1>
      <ul>
        {mons.map(mon => (
          <li key={mon.address}><a href="#" onClick={setMon.bind(null, mon.address)}>{mon.user} ({mon.host} - {mon.address})</a></li>
        ))}
      </ul>
      <h2>Active Monitor ({((mons.find(m => m.address == selected)) || { user: "" }).user}) {selected && <button onClick={setMon.bind(null, null)}>(Clear)</button>}</h2>
      <img width="900" src={imgSrc} />
    </div>
  );
}

export default App;
