'use client';

import { useState } from 'react';
import { uploadReport } from '@/lib/api';
import { PolicyReport } from '@/types/report';
import { Upload } from 'lucide-react';

interface Props {
  onReportLoad: (report: PolicyReport) => void;
}

export default function FileUpload({ onReportLoad }: Props) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');

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

  return (
    <div className="bg-white rounded-lg shadow p-6 mb-6">
      <h2 className="text-xl font-semibold mb-4">Load Report</h2>
      
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
            file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100
            disabled:opacity-50"
        />
      </div>

      {error && (
        <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
          {error}
        </div>
      )}
      
      {loading && (
        <div className="text-sm text-gray-600">Loading...</div>
      )}
    </div>
  );
}
