"use client"

import React from "react";
import { useState, useEffect, useRef } from "react";
import Axios from "axios";

export default function Dashboard() {

    const [spreadData, setSpreadData] = useState({
        niftyLTP: 0,
        shortLTP: 0,
        longLTP: 0,
        netSpread: 0,
        initialSpread: 0,
        status: "Waiting for feed..."
    })

    const [wsStatus, setWsStatus] = useState("Connecting to Live Feed...");
    const [positions, setPositions] = useState(null);
    const [isFetching, setIsFetching] = useState(false);
    
    const [selectedShort, setSelectedShort] = useState(null);
    const [selectedLong, setSelectedLong] = useState(null);
 
    const wsRef = useRef(null);

    useEffect(() => {
        wsRef.current = new WebSocket("ws://localhost:8080/ws");

        wsRef.current.onopen = () => {
            setWsStatus("Live Market Data Active");
        };

        wsRef.current.onmessage = (event) => {
            const data = JSON.parse(event.data);
            setSpreadData(data);
        };

        wsRef.current.onclose = () => {
            setWsStatus("Disconnected from Server");
        };

        return () => {
            if (wsRef.current) wsRef.current.close();
        };
    }, [])

    const fetchPositions = async () => {
        setIsFetching(true);
        try {
            const res = await Axios.get("http://localhost:8080/api/positions");
            setPositions(res.data);
        } catch (err) {
            console.error(err);
            alert("Error fetching positions. Make sure you are logged in.");
        } finally {
            setIsFetching(false);
        }
    };

    const handleTrackSpread = async () => {
        if(!selectedShort || !selectedLong) {
            alert("Please select both a Short Leg and a Long Leg first.");
            return;
        }

        try {
            const res = await Axios.post("http://localhost:8080/api/track-spread", {
                short_token: Number(selectedShort),
                long_token: Number(selectedLong)
            });

            if(res.data.status === "success") {
                alert("Backend is now tracking the spread!")
            }
        } catch (err) {
            console.error(err);
            alert("Failed to start tracking");
        }
    }

    return (
        
        <div>
            <div>
                <h1>Theta Harvester</h1>
                <div>{wsStatus}</div>
            </div>

            <div>
                <div>
                    <h3>NIFTY 50</h3>
                    <div>
                        ₹{spreadData.niftyLTP > 0 ? spreadData.niftyLTP.toFixed(2) : "-----.--"}
                    </div>
                </div>

                <div>
                    <div>
                        <h4>Short Leg LTP</h4>
                        <div>₹{spreadData.shortLTP > 0 ? spreadData.shortLTP.toFixed(2) : "0.00"}</div>
                    </div>
                    <div>
                        <h4>Long Leg LTP</h4>
                        <div>₹{spreadData.longLTP > 0 ? spreadData.longLTP.toFixed(2) : "0.00"}</div>
                    </div>
                    <div>
                        <h4>Collected Premium</h4>
                        <div>₹{spreadData.initialSpread > 0 ? spreadData.initialSpread.toFixed(2) : "0.00"}</div>
                    </div>
                </div>

                <div>
                    <h4>Live Net Spread</h4>
                    <div>₹{spreadData.netSpread > 0 ? spreadData.netSpread.toFixed(2) : "0.00"}</div>
                </div>
            </div>

            <div>
                <div>
                    <h4>Live Positions</h4>
                    <div>
                        <button onClick={fetchPositions} disabled={isFetching}>
                            {isFetching ? "Refreshing..." : "Refresh Positions"}
                        </button>
                        <button onClick={handleTrackSpread}>
                            Start Tracking Spread
                        </button>
                    </div>
                </div>

                {!positions ? (
                    <div>
                        Click "Refresh Positions" to pull your latest portfolio.
                    </div>
                ) : (!positions.net || positions.net.length === 0) ? (
                    <div>
                        No open positions found.
                    </div>
                ) : (
                    <table>
                        <thead>
                            <tr>
                                <th>Select Short</th>
                                <th>Select Long</th>
                                <th>Instrument</th>
                                <th>Qty</th>
                                <th>Avg Entry</th>
                                <th>P&L</th>
                            </tr>
                        </thead>
                        <tbody>
                            {positions.net?.map((pos, idx) => (
                                <tr key={idx}>
                                    <td>
                                        <input 
                                            type="radio" 
                                            name="shortLeg" 
                                            value={pos.instrument_token} 
                                            onChange={() => setSelectedShort(pos.instrument_token)} 
                                        />
                                    </td>
                                    <td>
                                        <input 
                                            type="radio" 
                                            name="longLeg" 
                                            value={pos.instrument_token} 
                                            onChange={() => setSelectedLong(pos.instrument_token)} 
                                        />
                                    </td>
                                    <td>{pos.tradingsymbol}</td>
                                    <td>{pos.quantity}</td>
                                    <td>₹{pos.average_price}</td>
                                    <td>₹{pos.pnl}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>
        </div>
    )
}