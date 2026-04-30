import Card from '../../components/Card';

export default function DashboardPage() {
  return (
    <>
        <h1>Dashboard</h1>
        {/* A11Y ISSUE: ensure proper heading order by starting with H1, followed by H2 */}
        <h2>Dashboard Overview</h2>
      <div className="grid">
        <Card title="Users" value="1,240" />
        <Card title="Revenue" value="$48k" />
        <Card title="Sessions" value="8,920" />
      </div>
    </>
  );
}
