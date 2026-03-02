import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import Navbar from './Navbar';
import Sidebar from './Sidebar';
import styles from './Layout.module.css';

export default function Layout() {
  const [sidebarOpen, setSidebarOpen] = useState(true);

  return (
    <div className={styles.root}>
      <Navbar
        sidebarOpen={sidebarOpen}
        onToggleSidebar={() => setSidebarOpen((prev) => !prev)}
      />
      <Sidebar open={sidebarOpen} />
      <main className={`${styles.main} ${sidebarOpen ? styles.mainShifted : styles.mainFull}`}>
        <Outlet />
      </main>
    </div>
  );
}