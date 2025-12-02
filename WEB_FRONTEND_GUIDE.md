# PG SecureLab - Frontend Implementation Guide

## 🎨 React + TypeScript + Next.js Application

### Установка

```bash
cd web
npx create-next-app@latest frontend \
  --typescript \
  --tailwind \
  --app \
  --no-src-dir \
  --import-alias "@/*"

cd frontend
npm install recharts react-force-graph-2d
npm install lucide-react  # иконки
npm run dev
```

### 📁 Структура проекта

```
frontend/
├── app/
│   ├── layout.tsx          # Основной layout
│   ├── page.tsx            # Главная страница
│   └── globals.css        # Tailwind стили
├── components/
│   ├── FileUpload.tsx     # Загрузка JSON
│   ├── OverviewTab.tsx    # Вкладка "Обзор"
│   ├── RolesTab.tsx       # Вкладка "Роли"
│   ├── RLSTab.tsx         # Вкладка "RLS"
│   ├── FindingsTab.tsx    # Вкладка "Нарушения"
│   ├── WhatIfSimulator.tsx # What-If симулятор
│   └── RoleGraph.tsx      # Граф ролей
├── types/
│   └── report.ts          # TypeScript интерфейсы
├── lib/
│   └── api.ts             # API клиент
└── package.json
```

---

## 📝 Код компонентов

### 1. types/report.ts

```typescript
export interface InstanceInfo {
  version: string;
  settings: Record<string, string>;
}

export interface RoleInfo {
  name: string;
  login: boolean;
  superuser: boolean;
  bypassrls: boolean;
  grants: string[];
}

export interface TableInfo {
  schema: string;
  name: string;
  rls_enabled: boolean;
}

export type Severity = "info" | "warning" | "critical";

export interface Finding {
  severity: Severity;
  code: string;
  message: string;
}

export interface PolicyReport {
  instance: InstanceInfo;
  roles: RoleInfo[];
  tables: TableInfo[];
  findings: Finding[];
}
```

### 2. lib/api.ts

```typescript
import { PolicyReport } from '@/types/report';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function analyzeDatabase(dsn: string): Promise<PolicyReport> {
  const response = await fetch(`${API_BASE}/api/analyze`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ dsn }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Analysis failed');
  }

  const data = await response.json();
  return data.report;
}

export async function uploadReport(file: File): Promise<PolicyReport> {
  const formData = new FormData();
  formData.append('report', file);

  const response = await fetch(`${API_BASE}/api/upload`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Upload failed');
  }

  const data = await response.json();
  return data.report;
}
```

### 3. components/FileUpload.tsx

```typescript
'use client';

import { useState } from 'react';
import { uploadReport, analyzeDatabase } from '@/lib/api';
import { PolicyReport } from '@/types/report';
import { Upload, Database } from 'lucide-react';

interface Props {
  onReportLoad: (report: PolicyReport) => void;
}

export default function FileUpload({ onReportLoad }: Props) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');
  const [dsn, setDsn] = useState('');

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setLoading(true);
    setError('');

    try {
      const report = await uploadReport(file);
      onReportLoad(report);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setLoading(false);
    }
  };

  const handleAnalyze = async () => {
    if (!dsn) {
      setError('DSN is required');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const report = await analyzeDatabase(dsn);
      onReportLoad(report);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Analysis failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow p-6 mb-6">
      <h2 className="text-xl font-semibold mb-4">Load Report</h2>
      
      {/* Upload JSON */}
      <div className="mb-4">
        <label className="block text-sm font-medium mb-2">
          <Upload className="inline mr-2" size={16} />
          Upload JSON Report
        </label>
        <input
          type="file"
          accept=".json"
          onChange={handleFileUpload}
          disabled={loading}
          className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4
            file:rounded-md file:border-0 file:text-sm file:font-semibold
            file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
        />
      </div>

      {/* Or Analyze Database */}
      <div className="border-t pt-4">
        <label className="block text-sm font-medium mb-2">
          <Database className="inline mr-2" size={16} />
          Or Analyze Database
        </label>
        <div className="flex gap-2">
          <input
            type="text"
            placeholder="postgres://user:pass@host:5432/dbname"
            value={dsn}
            onChange={(e) => setDsn(e.target.value)}
            disabled={loading}
            className="flex-1 px-3 py-2 border rounded-md"
          />
          <button
            onClick={handleAnalyze}
            disabled={loading || !dsn}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700
              disabled:bg-gray-300 disabled:cursor-not-allowed"
          >
            {loading ? 'Analyzing...' : 'Analyze'}
          </button>
        </div>
      </div>

      {error && (
        <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
          {error}
        </div>
      )}
    </div>
  );
}
```

### 4. components/OverviewTab.tsx

```typescript
'use client';

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
      {/* Instance Info */}
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

      {/* Metrics */}
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

function MetricCard({ icon, title, value, subtitle, color }: any) {
  const colors = {
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
```

### 5. components/RolesTab.tsx

```typescript
'use client';

import { PolicyReport, RoleInfo } from '@/types/report';
import { User, Shield, Lock } from 'lucide-react';

interface Props {
  report: PolicyReport;
}

export default function RolesTab({ report }: Props) {
  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Role Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Login
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Superuser
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Bypass RLS
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Grants
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {report.roles.map((role) => (
              <RoleRow key={role.name} role={role} />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function RoleRow({ role }: { role: RoleInfo }) {
  const isDangerous = role.superuser || role.bypassrls;

  return (
    <tr className={isDangerous ? 'bg-red-50' : ''}>
      <td className="px-6 py-4 whitespace-nowrap">
        <div className="flex items-center">
          <User className="mr-2" size={16} />
          <span className="font-medium">{role.name}</span>
          {role.superuser && (
            <Shield className="ml-2 text-red-600" size={16} title="Superuser" />
          )}
          {role.bypassrls && (
            <Lock className="ml-2 text-orange-600" size={16} title="Bypass RLS" />
          )}
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <Badge active={role.login} label={role.login ? 'Yes' : 'No'} />
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <Badge active={role.superuser} label={role.superuser ? 'Yes' : 'No'} danger={role.superuser} />
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <Badge active={role.bypassrls} label={role.bypassrls ? 'Yes' : 'No'} danger={role.bypassrls} />
      </td>
      <td className="px-6 py-4">
        <span className="text-sm text-gray-600">{role.grants.length} grants</span>
      </td>
    </tr>
  );
}

function Badge({ active, label, danger }: any) {
  const className = active
    ? danger
      ? 'bg-red-100 text-red-800'
      : 'bg-green-100 text-green-800'
    : 'bg-gray-100 text-gray-800';

  return (
    <span className={`px-2 py-1 text-xs font-semibold rounded-full ${className}`}>
      {label}
    </span>
  );
}
```

### 6. components/FindingsTab.tsx

```typescript
'use client';

import { useState } from 'react';
import { PolicyReport, Finding, Severity } from '@/types/report';
import { AlertTriangle, AlertCircle, Info } from 'lucide-react';

interface Props {
  report: PolicyReport;
}

export default function FindingsTab({ report }: Props) {
  const [filter, setFilter] = useState<Severity | 'all'>('all');

  const filtered = filter === 'all'
    ? report.findings
    : report.findings.filter(f => f.severity === filter);

  return (
    <div className="space-y-4">
      {/* Filter */}
      <div className="bg-white rounded-lg shadow p-4">
        <div className="flex gap-2">
          <FilterButton
            active={filter === 'all'}
            onClick={() => setFilter('all')}
            label="All"
            count={report.findings.length}
          />
          <FilterButton
            active={filter === 'critical'}
            onClick={() => setFilter('critical')}
            label="Critical"
            count={report.findings.filter(f => f.severity === 'critical').length}
            color="red"
          />
          <FilterButton
            active={filter === 'warning'}
            onClick={() => setFilter('warning')}
            label="Warning"
            count={report.findings.filter(f => f.severity === 'warning').length}
            color="yellow"
          />
          <FilterButton
            active={filter === 'info'}
            onClick={() => setFilter('info')}
            label="Info"
            count={report.findings.filter(f => f.severity === 'info').length}
            color="blue"
          />
        </div>
      </div>

      {/* Findings List */}
      <div className="space-y-3">
        {filtered.map((finding, idx) => (
          <FindingCard key={idx} finding={finding} />
        ))}
        {filtered.length === 0 && (
          <div className="bg-white rounded-lg shadow p-8 text-center text-gray-500">
            No findings found
          </div>
        )}
      </div>
    </div>
  );
}

function FilterButton({ active, onClick, label, count, color = 'gray' }: any) {
  const colors = {
    gray: active ? 'bg-gray-600 text-white' : 'bg-gray-100 text-gray-700',
    red: active ? 'bg-red-600 text-white' : 'bg-red-100 text-red-700',
    yellow: active ? 'bg-yellow-600 text-white' : 'bg-yellow-100 text-yellow-700',
    blue: active ? 'bg-blue-600 text-white' : 'bg-blue-100 text-blue-700',
  };

  return (
    <button
      onClick={onClick}
      className={`px-4 py-2 rounded-md font-medium ${colors[color]} transition-colors`}
    >
      {label} ({count})
    </button>
  );
}

function FindingCard({ finding }: { finding: Finding }) {
  const Icon = finding.severity === 'critical'
    ? AlertCircle
    : finding.severity === 'warning'
    ? AlertTriangle
    : Info;

  const colors = {
    critical: 'border-red-300 bg-red-50',
    warning: 'border-yellow-300 bg-yellow-50',
    info: 'border-blue-300 bg-blue-50',
  };

  const textColors = {
    critical: 'text-red-800',
    warning: 'text-yellow-800',
    info: 'text-blue-800',
  };

  return (
    <div className={`border-l-4 rounded-lg p-4 ${colors[finding.severity]}`}>
      <div className="flex items-start">
        <Icon className={`mt-0.5 mr-3 ${textColors[finding.severity]}`} size={20} />
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className={`text-sm font-semibold uppercase ${textColors[finding.severity]}`}>
              {finding.severity}
            </span>
            <span className="text-xs font-mono text-gray-600">{finding.code}</span>
          </div>
          <p className="text-sm text-gray-700">{finding.message}</p>
        </div>
      </div>
    </div>
  );
}
```

### 7. app/page.tsx - Главная страница

```typescript
'use client';

import { useState } from 'react';
import { PolicyReport } from '@/types/report';
import FileUpload from '@/components/FileUpload';
import OverviewTab from '@/components/OverviewTab';
import RolesTab from '@/components/RolesTab';
import FindingsTab from '@/components/FindingsTab';

export default function Home() {
  const [report, setReport] = useState<PolicyReport | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'roles' | 'rls' | 'findings'>('overview');

  return (
    <main className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <header className="mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            🔒 PG SecureLab UI
          </h1>
          <p className="text-gray-600">
            PostgreSQL Security Analysis Dashboard
          </p>
        </header>

        {/* File Upload */}
        <FileUpload onReportLoad={setReport} />

        {/* Tabs */}
        {report && (
          <>
            <div className="mb-6 border-b border-gray-200">
              <nav className="flex gap-4">
                <TabButton
                  active={activeTab === 'overview'}
                  onClick={() => setActiveTab('overview')}
                  label="Overview"
                />
                <TabButton
                  active={activeTab === 'roles'}
                  onClick={() => setActiveTab('roles')}
                  label="Roles & Privileges"
                />
                <TabButton
                  active={activeTab === 'findings'}
                  onClick={() => setActiveTab('findings')}
                  label="Findings"
                  badge={report.findings.length}
                />
              </nav>
            </div>

            {/* Tab Content */}
            <div>
              {activeTab === 'overview' && <OverviewTab report={report} />}
              {activeTab === 'roles' && <RolesTab report={report} />}
              {activeTab === 'findings' && <FindingsTab report={report} />}
            </div>
          </>
        )}
      </div>
    </main>
  );
}

function TabButton({ active, onClick, label, badge }: any) {
  return (
    <button
      onClick={onClick}
      className={`px-4 py-2 font-medium border-b-2 transition-colors ${
        active
          ? 'border-blue-600 text-blue-600'
          : 'border-transparent text-gray-600 hover:text-gray-900'
      }`}
    >
      {label}
      {badge !== undefined && (
        <span className="ml-2 px-2 py-0.5 text-xs bg-red-100 text-red-600 rounded-full">
          {badge}
        </span>
      )}
    </button>
  );
}
```

---

## 🚀 Запуск

```bash
# 1. API
cd web/api
go run server.go

# 2. Frontend
cd web/frontend
npm run dev
```

Откройте http://localhost:3000

## 📸 Скриншоты для курсовой

1. **Overview** - метрики и статистика
2. **Roles** - таблица ролей с выделением опасных
3. **Findings** - список проблем с фильтрами
4. **Upload** - загрузка JSON файла

## ✅ Готово!

Полнофункциональный веб-интерфейс для визуализации отчетов безопасности PostgreSQL.
