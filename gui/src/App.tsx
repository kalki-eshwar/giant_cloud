import React, { useEffect, useState } from 'react';
import './App.css';

interface NodeStatus {
  id: string;
  capacity: number;
  status: string;
}

function App() {
  const [nodes, setNodes] = useState<NodeStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchNodes = async () => {
    try {
      const response = await fetch('http://localhost:8081/api/nodes');
      if (!response.ok) throw new Error('Failed to fetch nodes');
      const data = await response.json();
      setNodes(data || []);
      setError(null);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNodes();
    const interval = setInterval(fetchNodes, 5000);
    return () => clearInterval(interval);
  }, []);

  const handlePing = async (nodeId: string) => {
    try {
      const response = await fetch(`http://localhost:8081/api/nodes/${encodeURIComponent(nodeId)}/ping`, {
        method: 'POST',
      });
      if (!response.ok) {
        const errData = await response.json();
        alert(`Error: ${errData.status || response.statusText}`);
      } else {
        // Optimistically show a success (or toast)
        console.log(`Pinged node ${nodeId}`);
      }
    } catch (err: any) {
      alert(`Ping failed: ${err.message}`);
    }
  };

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <div className="logo-container">
          <div className="logo-mark"></div>
          <h1>Giant Cloud Control Plane</h1>
        </div>
        <button className="refresh-button" onClick={fetchNodes}>
          <span className="icon">↻</span> Refresh
        </button>
      </header>

      <main className="dashboard-content">
        <section className="overview-panel">
          <h2>Network Overview</h2>
          <div className="stats-grid">
            <div className="stat-card">
              <div className="stat-value">{nodes.length}</div>
              <div className="stat-label">Total Nodes</div>
            </div>
            <div className="stat-card">
              <div className="stat-value active-count">
                {nodes.filter(n => n.status === 'active').length}
              </div>
              <div className="stat-label">Active Nodes</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">
                {nodes.reduce((sum, n) => sum + n.capacity, 0)} TB
              </div>
              <div className="stat-label">Total Capacity</div>
            </div>
          </div>
        </section>

        <section className="nodes-panel">
          <h2>Worker Nodes</h2>
          {error && <div className="error-banner">Warning: {error}. Is the Control Plane running on port 8081?</div>}
          
          {loading && nodes.length === 0 ? (
            <div className="loading-state">Connecting to network...</div>
          ) : (
            <div className="nodes-table-container">
              <table className="nodes-table">
                <thead>
                  <tr>
                    <th>Status</th>
                    <th>Node Address</th>
                    <th>Capacity</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {nodes.length === 0 && !loading && (
                    <tr>
                      <td colSpan={4} className="empty-state">No nodes registered to the network.</td>
                    </tr>
                  )}
                  {nodes.map((node) => (
                    <tr key={node.id} className={node.status === 'offline' ? 'offline-row' : ''}>
                      <td>
                        <div className={`status-badge ${node.status}`}>
                          <span className="status-dot"></span>
                          {node.status}
                        </div>
                      </td>
                      <td className="node-id">{node.id}</td>
                      <td>{node.capacity} TB</td>
                      <td>
                        <button 
                          className="action-button" 
                          onClick={() => handlePing(node.id)}
                          disabled={node.status === 'offline'}
                        >
                          Ping Proof
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </main>
    </div>
  );
}

export default App;
