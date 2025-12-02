'use client';

import type { ReactNode } from 'react';
import { PolicyReport } from '@/types/report';
import { Server, Shield, Table, AlertTriangle } from 'lucide-react';

interface Props {
  report: PolicyReport;
}

export default function OverviewTab({ report }: Props) {
  const rlsEnabled = report.tables.filter(t => t.rls_enabled).length;
  const rlsDisabled = report.tables.length - rlsEnabled;
  const criticalFindings = report.findings.filter(f => f.severity === 'critical').length;
  const warningFindings = report.findings.filter(f => f.severity === 'warning').length;

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold mb-4 flex items-center">
          <Server className="mr-2" size={20} />
          PostgreSQL Instance
        </h3>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <span className="text-sm text-gray-600">Version:</span>
            <p className="font-mono text-sm">{report.instance.version}</p>
          </div>
          {Object.entries(report.instance.settings).map(([key, value]) => (
            <div key={key}>
              <span className="text-sm text-gray-600">{key}:</span>
              <p className="font-mono text-sm">{value}</p>
            </div>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          icon={<Shield size={24} />}
          title="Roles"
          value={report.roles.length}
          color="blue"
        />
        <MetricCard
          icon={<Table size={24} />}
          title="Tables"
          value={report.tables.length}
          color="green"
        />
        <MetricCard
          icon={<Shield size={24} />}
          title="RLS Enabled"
          value={rlsEnabled}
          subtitle={`${rlsDisabled} disabled`}
          color="green"
        />
        <MetricCard
          icon={<AlertTriangle size={24} />}
          title="Findings"
          value={report.findings.length}
          subtitle={`${criticalFindings} critical, ${warningFindings} warning`}
          color="red"
        />
      </div>
    </div>
  );
}
type Color = 'blue' | 'green' | 'red';

interface MetricCardProps {
  icon: ReactNode;
  title: string;
  value: number | string;
  subtitle?: string;
  color: Color;
}

function MetricCard({ icon, title, value, subtitle, color }: MetricCardProps) {
  const colors: Record<Color, string> = {
    blue: 'bg-blue-50 text-blue-600',
    green: 'bg-green-50 text-green-600',
    red: 'bg-red-50 text-red-600',
  };

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className={`inline-flex p-3 rounded-lg ${colors[color]} mb-3`}>
        {icon}
      </div>
      <h4 className="text-sm font-medium text-gray-600">{title}</h4>
      <p className="text-2xl font-bold mt-1">{value}</p>
      {subtitle && <p className="text-xs text-gray-500 mt-1">{subtitle}</p>}
    </div>
  );
}

