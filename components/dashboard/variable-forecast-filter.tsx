"use client";

export function VariableForecastFilter({ include, onChange }: { include: boolean; onChange: (include: boolean) => void }) {
  return <label className="native-field"><span>Previsões variáveis</span>
    <select value={String(include)} onChange={(event) => onChange(event.target.value === "true")}>
      <option value="false">Sem previsões variáveis</option>
      <option value="true">Com previsões variáveis</option>
    </select>
    <small>As cobranças importadas entram nas duas visões.</small>
  </label>;
}
