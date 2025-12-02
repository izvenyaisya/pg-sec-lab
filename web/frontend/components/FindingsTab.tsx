'use client';

import { useState } from 'react';
import { PolicyReport, Finding, Severity } from '@/types/report';
import { AlertTriangle, AlertCircle, Info } from 'lucide-react';

interface Props {
  report: PolicyReport;
}

export default function FindingsTab({ report }: Props) {
  const [filter, setFilter] = useState<Severity | 'all'>('all');

  const filtered = filter === 'all' ? report.findings : report.findings.filter(f => f.severity === filter);

  return (
    <div className="space-y-4">
      <div className="bg-white rounded-lg shadow p-4 flex gap-2">
        <FilterBtn active={filter === 'all'} onClick={() => setFilter('all')} label="All" count={report.findings.length} />
        <FilterBtn active={filter === 'critical'} onClick={() => setFilter('critical')} label="Critical" count={report.findings.filter(f => f.severity === 'critical').length} color="red" />
        <FilterBtn active={filter === 'warning'} onClick={() => setFilter('warning')} label="Warning" count={report.findings.filter(f => f.severity === 'warning').length} color="yellow" />
      </div>

      <div className="space-y-3">
        {filtered.map((finding, idx) => (
          <FindingCard key={idx} finding={finding} />
        ))}
        {filtered.length === 0 && (
          <div className="bg-white rounded-lg shadow p-8 text-center text-gray-500">No findings</div>
        )}
      </div>
    </div>
  );
}

function FilterBtn({ active, onClick, label, count, color = 'gray' }: any) {
  const colors: any = {
    gray: active ? 'bg-gray-600 text-white' : 'bg-gray-100 text-gray-700',
    red: active ? 'bg-red-600 text-white' : 'bg-red-100 text-red-700',
    yellow: active ? 'bg-yellow-600 text-white' : 'bg-yellow-100 text-yellow-700',
  };

  return <button onClick={onClick} className={`px-4 py-2 rounded-md font-medium ${colors[color]}`}>{label} ({count})</button>;
}

function FindingCard({ finding }: { finding: Finding }) {
  const Icon = finding.severity === 'critical' ? AlertCircle : finding.severity === 'warning' ? AlertTriangle : Info;
  const colors: any = {
    critical: 'border-red-300 bg-red-50 text-red-800',
    warning: 'border-yellow-300 bg-yellow-50 text-yellow-800',
    info: 'border-blue-300 bg-blue-50 text-blue-800',
  };

  return (
    <div className={`border-l-4 rounded-lg p-4 ${colors[finding.severity]}`}>
      <div className="flex items-start">
        <Icon className="mt-0.5 mr-3" size={20} />
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm font-semibold uppercase">{finding.severity}</span>
            <span className="text-xs font-mono text-gray-600">{finding.code}</span>
          </div>
          <p className="text-sm">{finding.message}</p>
        </div>
      </div>
    </div>
  );
}
