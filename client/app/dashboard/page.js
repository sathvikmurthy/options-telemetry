"use client"

import React from "react";
import { useState, useEffect, useRef } from "react";
import Axios from "axios";
import "./dashboard.css";

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
        <div className="dashboard-container">
            
            {/* Header */}
            <header className="dashboard-header">
                <h1 className="dashboard-title">Theta Harvester</h1>
                <div 
                    className="status-indicator" 
                    style={{ color: wsStatus.includes("🟢") ? "#10b981" : "#ef4444" }}
                >
                    <span 
                        className="status-dot" 
                        style={{ backgroundColor: wsStatus.includes("🟢") ? "#10b981" : "#ef4444" }}
                    ></span>
                    {wsStatus}
                </div>
            </header>

            {/* Top Widgets */}
            <div className="widgets-row">
                
                {/* Nifty 50 Card */}
                <div className="terminal-card nifty-card">
                    <h3 className="card-label">NIFTY 50</h3>
                    <div className="card-value" style={{ color: "#eab308" }}>
                        <span>₹{spreadData.niftyLTP > 0 ? spreadData.niftyLTP.toFixed(2) : "----.--"}</span>
                    </div>
                </div>

                {/* Spread Monitor */}
                <div className="terminal-card spread-card">
                    <div className="spread-legs">
                        <div>
                            <h4 className="card-label">Short Leg LTP</h4>
                            <div className="card-value-small">
                                <span>₹{spreadData.shortLTP > 0 ? spreadData.shortLTP.toFixed(2) : "0.00"}</span>
                            </div>
                        </div>
                        <div>
                            <h4 className="card-label">Long Leg LTP</h4>
                            <div className="card-value-small">
                                <span>₹{spreadData.longLTP > 0 ? spreadData.longLTP.toFixed(2) : "0.00"}</span>
                            </div>
                        </div>
                        <div className="divider">
                            <h4 className="card-label">Collected Premium</h4>
                            <div className="card-value-muted">
                                <span>₹{spreadData.initialSpread > 0 ? spreadData.initialSpread.toFixed(2) : "0.00"}</span>
                            </div>
                        </div>
                    </div>

                    <div className="net-spread-section">
                        <h4 className="card-label" style={{ marginBottom: "8px" }}>Live Net Spread</h4>
                        <div 
                            className="card-value-large" 
                            style={{ color: spreadData.netSpread > 0 ? "#3b82f6" : "#fafafa" }}
                        >
                            <span>₹{spreadData.netSpread > 0 ? spreadData.netSpread.toFixed(2) : "0.00"}</span>
                        </div>
                    </div>
                </div>
            </div>

            {/* Positions Table Container */}
            <div className="terminal-card">
                <div className="table-header">
                    <h4 className="table-title">Live Positions</h4>
                    <div className="button-group">
                        <button 
                            className="btn btn-secondary"
                            onClick={fetchPositions} 
                            disabled={isFetching}
                        >
                            {isFetching ? "Refreshing..." : "Refresh Positions"}
                        </button>
                        <button 
                            className="btn btn-primary"
                            onClick={handleTrackSpread}
                        >
                            Start Tracking Spread
                        </button>
                    </div>
                </div>

                {!positions ? (
                    <div className="empty-state">
                        Click "Refresh Positions" to pull your latest portfolio.
                    </div>
                ) : (!positions.net || positions.net.length === 0) ? (
                    <div className="empty-state">
                        No open positions found.
                    </div>
                ) : (
                    <div className="table-responsive">
                        <table className="positions-table">
                            <thead>
                                <tr>
                                    <th className="text-center" style={{ width: '80px' }}>Short</th>
                                    <th className="text-center" style={{ width: '80px' }}>Long</th>
                                    <th className="text-left">Instrument</th>
                                    <th className="text-right">Qty</th>
                                    <th className="text-right">Avg Entry</th>
                                    <th className="text-right">P&L</th>
                                </tr>
                            </thead>
                            <tbody>
                                {positions.net?.map((pos, idx) => (
                                    <tr key={idx}>
                                        <td className="text-center">
                                            <input 
                                                type="radio" 
                                                name="shortLeg" 
                                                value={pos.instrument_token} 
                                                onChange={() => setSelectedShort(pos.instrument_token)} 
                                                className="radio-short"
                                            />
                                        </td>
                                        <td className="text-center">
                                            <input 
                                                type="radio" 
                                                name="longLeg" 
                                                value={pos.instrument_token} 
                                                onChange={() => setSelectedLong(pos.instrument_token)} 
                                                className="radio-long"
                                            />
                                        </td>
                                        <td className="text-left" style={{ fontWeight: '600', color: '#fafafa' }}>
                                            {pos.tradingsymbol}
                                        </td>
                                        <td 
                                            className="text-right font-mono" 
                                            style={{ color: pos.quantity > 0 ? "#10b981" : (pos.quantity < 0 ? "#ef4444" : "#a1a1aa") }}
                                        >
                                            {pos.quantity}
                                        </td>
                                        <td className="text-right font-mono" style={{ color: '#a1a1aa' }}>
                                            ₹{pos.average_price}
                                        </td>
                                        <td 
                                            className="text-right font-mono"
                                            style={{ color: pos.pnl >= 0 ? "#10b981" : "#ef4444" }}
                                        >
                                            ₹{pos.pnl}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </div>
    )
}