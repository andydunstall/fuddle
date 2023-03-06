import React from 'react';

import Logo from './logo.png';

import './Nav.css';

export default function Nav() {
  return (
    <div className="nav">
      <div className="title">
        <img src={Logo} alt="Logo" />
      </div>
    </div>
  );
}
