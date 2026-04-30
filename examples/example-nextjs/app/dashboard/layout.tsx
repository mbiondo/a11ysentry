export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="dashboard">
      <aside>
        <nav aria-label="Dashboard navigation">
          <a href="/dashboard">Overview</a>
          <a href="/dashboard/settings">Settings</a>
        </nav>
      </aside>
      <section>{children}</section>
    </div>
  );
}
