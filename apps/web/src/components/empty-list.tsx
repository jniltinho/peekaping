const EmptyList = ({
  title,
  text,
  actionText,
  onClick,
}: {
  title: string;
  text: string;
  actionText: string;
  onClick: () => void;
}) => {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <h3 className="text-lg font-semibold mb-2">{title}</h3>
      <p className="text-muted-foreground mb-6 max-w-sm">{text}</p>
      <button
        onClick={onClick}
        className="inline-flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-none hover:bg-primary/90 transition-colors"
      >
        {actionText}
      </button>
    </div>
  );
};

export default EmptyList;
