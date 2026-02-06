import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft, Download, Copy, Check } from 'lucide-react';
import { useSAMLMetadata } from '../hooks/useSAML';
import { toast } from '../services/toast';

const SAMLMetadata: React.FC = () => {
  const { data, isLoading, error } = useSAMLMetadata();
  const [copied, setCopied] = useState(false);

  const handleDownload = () => {
    if (!data?.metadata) return;

    const blob = new Blob([data.metadata], { type: 'application/xml' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'saml-metadata.xml';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleCopy = async () => {
    if (!data?.metadata) return;

    try {
      await navigator.clipboard.writeText(data.metadata);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      console.error('Failed to copy:', error);
      toast.error('Failed to copy metadata to clipboard');
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <p className="text-destructive">Error loading SAML metadata: {(error as Error).message}</p>
        <Link to="/saml" className="text-primary hover:underline mt-4 inline-block">
          Back to SAML Service Providers
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/saml" className="text-muted-foreground hover:text-foreground flex items-center gap-2">
            <ArrowLeft size={20} />
            Back
          </Link>
          <div>
            <h1 className="text-2xl font-bold text-foreground">SAML IdP Metadata</h1>
            <p className="text-muted-foreground mt-1">Share this metadata with Service Providers</p>
          </div>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleCopy}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors flex items-center gap-2"
          >
            {copied ? <Check size={16} className="text-success" /> : <Copy size={16} />}
            {copied ? 'Copied!' : 'Copy'}
          </button>
          <button
            onClick={handleDownload}
            className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors flex items-center gap-2"
          >
            <Download size={16} />
            Download
          </button>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="p-4 border-b border-border bg-muted">
          <h2 className="text-sm font-semibold text-foreground">Metadata XML</h2>
          <p className="text-xs text-muted-foreground mt-1">Copy this XML and provide it to Service Providers</p>
        </div>
        <div className="p-6">
          <pre className="bg-foreground text-background p-4 rounded-lg overflow-x-auto text-xs font-mono">
            {data?.metadata || 'No metadata available'}
          </pre>
        </div>
      </div>

      <div className="bg-primary/10 border border-primary/20 rounded-lg p-4">
        <h3 className="font-semibold text-primary mb-2">How to use this metadata</h3>
        <ol className="list-decimal list-inside space-y-1 text-sm text-primary">
          <li>Download or copy the metadata XML above</li>
          <li>Provide it to the Service Provider administrator</li>
          <li>They will import it into their SAML configuration</li>
          <li>Configure the Service Provider in this system</li>
        </ol>
      </div>
    </div>
  );
};

export default SAMLMetadata;

