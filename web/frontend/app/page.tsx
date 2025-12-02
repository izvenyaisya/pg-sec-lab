'use client';

import { useState } from 'react';
import { PolicyReport } from '@/types/report';
import FileUpload from '@/components/FileUpload';
import OverviewTab from '@/components/OverviewTab';
import RolesTab from '@/components/RolesTab';
import FindingsTab from '@/components/FindingsTab';

export default function Home() {
  const [report, setReport] = useState<PolicyReport | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'roles' | 'findings'>('overview');

  return (
    <main className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <header className="mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            🔒 PG SecureLab UI
          </h1>
          <p className="text-gray-600">
            PostgreSQL Security Analysis Dashboard
          </p>
        </header>

        <FileUpload onReportLoad={setReport} />

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
                  label="Roles"
                />
                <TabButton
                  active={activeTab === 'findings'}
                  onClick={() => setActiveTab('findings')}
                  label="Findings"
                  badge={report.findings.length}
                />
              </nav>
            </div>

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

