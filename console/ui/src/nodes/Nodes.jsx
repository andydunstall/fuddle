import React, { useEffect, useState } from 'react';

import './Nodes.css';

function trimRevision(revision) {
  if (revision?.length > 10) {
    return `${revision.substring(0, 10)}...`;
  }
  return revision;
}

export default function Nodes() {
  const [nodes, setNodes] = useState([]);

  useEffect(() => {
    fetch('/api/v1/cluster')
      .then((resp) => resp.json())
      .then((resp) => {
        setNodes(resp.sort((a, b) => {
          if (a.id > b.id) return 1;
          if (a.id < b.id) return -1;
          return 0;
        }));
      })
      .catch(console.error);
  }, []);

  return (
    <div className="nodes">
      <h2>Nodes</h2>
      <div className="list">
        <table>
          <tr id="header">
            <th>ID</th>
            <th>Service</th>
            <th>Locality</th>
            <th>Created</th>
            <th>Revision</th>
          </tr>
          {
            nodes.map((n) => (
              <tr id="node">
                <td>{n.id}</td>
                <td>{n.service}</td>
                <td>{n.locality}</td>
                <td>{new Date(n.created).toLocaleString() }</td>
                <td>{trimRevision(n.revision)}</td>
              </tr>
            ))
          }
        </table>
      </div>
    </div>
  );
}
