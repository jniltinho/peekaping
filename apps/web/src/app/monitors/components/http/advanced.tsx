import { MultiSelect } from "@/components/multi-select";
import {
  FormControl,
  FormField,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { FormItem } from "@/components/ui/form";
import { TypographyH4 } from "@/components/ui/typography";
import { useFormContext } from "react-hook-form";
import { z } from "zod";

const acceptedStatusCodesOptions = [
  { value: "1XX", label: "1XX" },
  { value: "2XX", label: "2XX" },
  { value: "3XX", label: "3XX" },
  { value: "4XX", label: "4XX" },
  { value: "5XX", label: "5XX" },
];

export const advancedSchema = z.object({
  accepted_statuscodes: z.array(z.string()),
});

export type AdvancedForm = z.infer<typeof advancedSchema>;

export const advancedDefaultValues: AdvancedForm = {
  accepted_statuscodes: ["2XX"],
}

const Advanced = () => {
  const form = useFormContext();

  return (
    <>
      <TypographyH4>Advanced</TypographyH4>

      <FormField
        control={form.control}
        name="accepted_statuscodes"
        render={({ field }) => {
          return <FormItem>
          <FormLabel>Accepted Status Codes</FormLabel>
          <FormControl>
            <MultiSelect
              options={acceptedStatusCodesOptions}
              onValueChange={(val) => {
                field.onChange(val)
              }}
              value={field.value || []}
            />
          </FormControl>
          <FormMessage />
        </FormItem>
        }}
      />
    </>
  );
};

export default Advanced;
