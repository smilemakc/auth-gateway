import React from 'react';
import { Upload, Download } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface BulkCreateCSVParserProps {
  csvText: string;
  mode: 'manual' | 'csv' | 'json';
  onFileUpload: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onCSVTextChange: (text: string) => void;
  onDownloadTemplate: () => void;
}

const BulkCreateCSVParser: React.FC<BulkCreateCSVParserProps> = ({
  csvText,
  mode,
  onFileUpload,
  onCSVTextChange,
  onDownloadTemplate,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-6">
      <h2 className="text-lg font-semibold text-foreground mb-4">{t('bulk.upload_file')}</h2>
      <div className="space-y-4">
        <div className="flex items-center gap-4">
          <label className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg cursor-pointer flex items-center gap-2">
            <Upload size={16} />
            {t('bulk.choose_file')}
            <input type="file" accept=".csv,.json" onChange={onFileUpload} className="hidden" />
          </label>
          <button
            onClick={onDownloadTemplate}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent flex items-center gap-2"
          >
            <Download size={16} />
            {t('bulk.download_template')}
          </button>
        </div>

        {mode === 'csv' && (
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('bulk.csv_content')}</label>
            <textarea
              value={csvText}
              onChange={(e) => onCSVTextChange(e.target.value)}
              rows={10}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring font-mono text-xs"
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default BulkCreateCSVParser;
