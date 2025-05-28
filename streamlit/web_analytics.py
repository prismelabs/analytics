import streamlit as st
import clickhouse_connect

client = clickhouse_connect.get_client(username="clickhouse", password="password", database="prisme")
domains = []
with client.query_row_block_stream("select distinct(domain) from prisme.sessions") as stream:
	for block in stream:
		for row in block:
			domains = domains + [row[0]]

print(domains)

selected_domains = st.multiselect("Domains", domains, default=domains)

sessions = client.command("select count(*) from prisme.sessions where domain in ({sessions})", parameters={"sessions": ', '.join(selected_domains)})

# st.write(f"domains: {domains}")
# st.write(f"sessions: {sessions}")
