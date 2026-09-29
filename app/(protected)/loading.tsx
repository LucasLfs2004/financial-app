export default function Loading() {
  return <main className="resource-page-shell page-loading" role="status" aria-label="Carregando página">
    <div className="page-loading-line short" />
    <div className="page-loading-line title" />
    <div className="page-loading-line" />
    <div className="page-loading-panels"><div /><div /><div /></div>
  </main>;
}
