import {
  getCurrentUserQueryKey,
  getSessionQueryKey,
} from "@asyncstatus/sdk/@tanstack/react-query.gen";
import { getSession } from "@asyncstatus/sdk/sdk.gen";
import { CurrentUser } from "@asyncstatus/sdk/types.gen";
import { queryOptions } from "@tanstack/react-query";

export function getAuthQueryKey() {
  return [getSessionQueryKey(), getCurrentUserQueryKey()];
}

export function authQueryOptions() {
  return queryOptions({
    throwOnError: true,
    staleTime: Infinity,
    queryKey: getAuthQueryKey(),
    queryFn: async ({ client, signal }) => {
      const session = await getSession({ signal });
      const currentUser = {
        id: session.data.user_id,
        name: session.data.user_name,
        email: session.data.user_email,
        avatar_url: session.data.user_avatar_url,
        timezone: session.data.user_timezone,
      } satisfies CurrentUser;
      client.setQueryData(getCurrentUserQueryKey(), currentUser);
      return currentUser;
    },
  });
}
